import Foundation
import Combine
import Darwin

public final class SystemMonitor: ObservableObject {
    public static let shared = SystemMonitor()

    @Published public var cpuUsage: Double = 0.0          // 0.0 - 100.0%
    @Published public var memoryUsage: Double = 0.0       // 0.0 - 100.0%
    @Published public var memoryUsedFormatted: String = "0 GB"
    @Published public var memoryTotalFormatted: String = "0 GB"
    
    @Published public var diskUsagePercent: Double = 0.0  // 0.0 - 100.0%
    @Published public var diskFreeFormatted: String = "0 GB"
    @Published public var diskTotalFormatted: String = "0 GB"

    @Published public var downloadSpeedBytesPerSec: Double = 0.0
    @Published public var uploadSpeedBytesPerSec: Double = 0.0
    @Published public var downloadSpeedFormatted: String = "0 KB/s"
    @Published public var uploadSpeedFormatted: String = "0 KB/s"

    // 历史网速数据队列（用于实时波形图绘制，保留最近 30 秒）
    @Published public var downloadHistory: [Double] = Array(repeating: 0.0, count: 25)
    @Published public var uploadHistory: [Double] = Array(repeating: 0.0, count: 25)

    private var timer: AnyCancellable?
    private var prevCpuInfo: processor_info_array_t?
    private var prevCpuInfoCount: mach_msg_type_number_t = 0
    private var prevNetworkInBytes: UInt64 = 0
    private var prevNetworkOutBytes: UInt64 = 0
    private var lastSampleTime: Date = Date()

    private init() {
        startMonitoring()
    }

    public func startMonitoring() {
        // 初始采样
        updateDisk()
        updateMemory()
        _ = sampleNetworkBytes()
        _ = sampleCpuUsage()

        timer = Timer.publish(every: 1.0, on: .main, in: .common)
            .autoconnect()
            .sink { [weak self] _ in
                self?.refresh()
            }
    }

    public func refresh() {
        let now = Date()
        let interval = max(0.1, now.timeIntervalSince(lastSampleTime))
        lastSampleTime = now

        updateCpu()
        updateMemory()
        updateDisk()
        updateNetwork(interval: interval)
    }

    // MARK: - CPU Usage
    private func updateCpu() {
        self.cpuUsage = sampleCpuUsage()
    }

    private func sampleCpuUsage() -> Double {
        var processorCount: natural_t = 0
        var processorInfo: processor_info_array_t?
        var processorInfoCount: mach_msg_type_number_t = 0

        let result = host_processor_info(
            mach_host_self(),
            PROCESSOR_CPU_LOAD_INFO,
            &processorCount,
            &processorInfo,
            &processorInfoCount
        )

        guard result == KERN_SUCCESS, let cpuInfo = processorInfo else {
            return 0.0
        }

        var totalUsage: Double = 0.0

        if let prevInfo = prevCpuInfo {
            for i in 0..<Int(processorCount) {
                let offset = Int(CPU_STATE_MAX) * i
                let user = Double(cpuInfo[offset + Int(CPU_STATE_USER)] - prevInfo[offset + Int(CPU_STATE_USER)])
                let system = Double(cpuInfo[offset + Int(CPU_STATE_SYSTEM)] - prevInfo[offset + Int(CPU_STATE_SYSTEM)])
                let idle = Double(cpuInfo[offset + Int(CPU_STATE_IDLE)] - prevInfo[offset + Int(CPU_STATE_IDLE)])
                let nice = Double(cpuInfo[offset + Int(CPU_STATE_NICE)] - prevInfo[offset + Int(CPU_STATE_NICE)])

                let total = user + system + idle + nice
                if total > 0 {
                    totalUsage += ((user + system + nice) / total)
                }
            }
            totalUsage = (totalUsage / Double(processorCount)) * 100.0
        }

        if let prev = prevCpuInfo {
            let prevSize = vm_size_t(prevCpuInfoCount) * vm_size_t(MemoryLayout<integer_t>.size)
            vm_deallocate(mach_task_self_, vm_address_t(bitPattern: prev), prevSize)
        }

        prevCpuInfo = cpuInfo
        prevCpuInfoCount = processorInfoCount

        return min(100.0, max(0.0, totalUsage))
    }

    // MARK: - Memory Usage
    private func updateMemory() {
        var size = mach_msg_type_number_t(MemoryLayout<vm_statistics64_data_t>.size / MemoryLayout<integer_t>.size)
        var vmStats = vm_statistics64()
        let hostPort = mach_host_self()

        let ret = withUnsafeMutablePointer(to: &vmStats) {
            $0.withMemoryRebound(to: integer_t.self, capacity: Int(size)) {
                host_statistics64(hostPort, HOST_VM_INFO64, $0, &size)
            }
        }

        guard ret == KERN_SUCCESS else { return }

        let pageSize = UInt64(vm_kernel_page_size)
        let active = UInt64(vmStats.active_count) * pageSize
        let wired = UInt64(vmStats.wire_count) * pageSize
        let compressed = UInt64(vmStats.compressor_page_count) * pageSize
        let used = active + wired + compressed

        let total = ProcessInfo.processInfo.physicalMemory

        let percent = Double(used) / Double(total) * 100.0
        self.memoryUsage = min(100.0, max(0.0, percent))
        self.memoryUsedFormatted = ByteFormatter.format(Int64(used))
        self.memoryTotalFormatted = ByteFormatter.format(Int64(total))
    }

    // MARK: - Disk Usage
    private func updateDisk() {
        let homeUrl = URL(fileURLWithPath: NSHomeDirectory())
        do {
            let values = try homeUrl.resourceValues(forKeys: [.volumeTotalCapacityKey, .volumeAvailableCapacityKey])
            if let total = values.volumeTotalCapacity, let available = values.volumeAvailableCapacity {
                let used = Int64(total - available)
                let percent = (Double(used) / Double(total)) * 100.0
                self.diskUsagePercent = min(100.0, max(0.0, percent))
                self.diskFreeFormatted = ByteFormatter.format(Int64(available))
                self.diskTotalFormatted = ByteFormatter.format(Int64(total))
            }
        } catch {
            // fallback
        }
    }

    // MARK: - Network Speed
    private func sampleNetworkBytes() -> (UInt64, UInt64) {
        var ifap: UnsafeMutablePointer<ifaddrs>?
        guard getifaddrs(&ifap) == 0, let first = ifap else { return (0, 0) }
        defer { freeifaddrs(ifap) }

        var inBytes: UInt64 = 0
        var outBytes: UInt64 = 0

        var cursor: UnsafeMutablePointer<ifaddrs>? = first
        while let current = cursor {
            let flags = Int32(current.pointee.ifa_flags)
            // Filter loopback and inactive
            if (flags & IFF_UP) != 0 && (flags & IFF_LOOPBACK) == 0 {
                let name = String(cString: current.pointee.ifa_name)
                // Filter primary interfaces like en0, en1, pdp_ip0, etc.
                if name.hasPrefix("en") || name.hasPrefix("ap") || name.hasPrefix("pdp_ip") {
                    if let data = current.pointee.ifa_data {
                        let ifData = data.assumingMemoryBound(to: if_data64.self)
                        inBytes += ifData.pointee.ifi_ibytes
                        outBytes += ifData.pointee.ifi_obytes
                    }
                }
            }
            cursor = current.pointee.ifa_next
        }

        return (inBytes, outBytes)
    }

    private func updateNetwork(interval: TimeInterval) {
        let (currentIn, currentOut) = sampleNetworkBytes()
        if prevNetworkInBytes > 0 && prevNetworkOutBytes > 0 && interval > 0 {
            let deltaIn = currentIn >= prevNetworkInBytes ? currentIn - prevNetworkInBytes : 0
            let deltaOut = currentOut >= prevNetworkOutBytes ? currentOut - prevNetworkOutBytes : 0

            let inSpeed = Double(deltaIn) / interval
            let outSpeed = Double(deltaOut) / interval

            self.downloadSpeedBytesPerSec = inSpeed
            self.uploadSpeedBytesPerSec = outSpeed

            self.downloadSpeedFormatted = formatSpeed(inSpeed)
            self.uploadSpeedFormatted = formatSpeed(outSpeed)

            // 更新折线历史队列
            var dl = self.downloadHistory
            dl.removeFirst()
            dl.append(inSpeed)
            self.downloadHistory = dl

            var ul = self.uploadHistory
            ul.removeFirst()
            ul.append(outSpeed)
            self.uploadHistory = ul
        }

        prevNetworkInBytes = currentIn
        prevNetworkOutBytes = currentOut
    }

    private func formatSpeed(_ bytesPerSec: Double) -> String {
        let kb = bytesPerSec / 1024
        let mb = kb / 1024
        if mb >= 1.0 {
            return String(format: "%.1f MB/s", mb)
        } else if kb >= 1.0 {
            return String(format: "%.0f KB/s", kb)
        } else {
            return String(format: "%.0f B/s", bytesPerSec)
        }
    }
}

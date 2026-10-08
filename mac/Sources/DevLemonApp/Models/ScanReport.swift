import Foundation
import SwiftUI

public enum RiskLevel: String, Codable, CaseIterable {
    case safe = "safe"
    case rebuildable = "rebuildable"
    case caution = "caution"

    public var title: String {
        switch self {
        case .safe: return "安全"
        case .rebuildable: return "可重构"
        case .caution: return "谨慎"
        }
    }

    public var color: Color {
        switch self {
        case .safe: return Color.green
        case .rebuildable: return Color.orange
        case .caution: return Color.red
        }
    }
}

public enum ItemCategory: String, Codable, CaseIterable {
    case docker = "docker"
    case simulator = "simulator"
    case packageCache = "package_cache"
    case workspaceBuild = "workspace_build"
    case systemCache = "system_cache"
    case aiCache = "ai_cache"

    public var displayName: String {
        switch self {
        case .docker: return "Docker 容器"
        case .simulator: return "模拟器沙盒"
        case .packageCache: return "包管理缓存"
        case .workspaceBuild: return "项目构建产物"
        case .systemCache: return "系统维护垃圾"
        case .aiCache: return "AI 模型与工具"
        }
    }

    public var iconName: String {
        switch self {
        case .docker: return "shippingbox.fill"
        case .simulator: return "iphone"
        case .packageCache: return "shippingbox"
        case .workspaceBuild: return "hammer.fill"
        case .systemCache: return "trash.fill"
        case .aiCache: return "brain.head.profile"
        }
    }
}

public struct Item: Codable, Identifiable, Hashable, Equatable {
    public var id: String
    public var title: String
    public var description: String
    public var path: String?
    public var sizeBytes: Int64
    public var sizeFormatted: String
    public var risk: RiskLevel
    public var category: ItemCategory
    public var isProtected: Bool
    public var protectReason: String?

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case description
        case path
        case sizeBytes = "size_bytes"
        case sizeFormatted = "size_formatted"
        case risk
        case category
        case isProtected = "is_protected"
        case protectReason = "protect_reason"
    }

    public static func == (lhs: Item, rhs: Item) -> Bool {
        lhs.id == rhs.id
    }

    public func hash(into hasher: inout Hasher) {
        hasher.combine(id)
    }
}

public struct Group: Codable, Identifiable {
    public var id: String
    public var title: String
    public var category: ItemCategory
    public var totalSizeBytes: Int64
    public var totalReclaimableBytes: Int64
    public var items: [Item]

    enum CodingKeys: String, CodingKey {
        case id
        case title
        case category
        case totalSizeBytes = "total_size_bytes"
        case totalReclaimableBytes = "total_reclaimable_bytes"
        case items
    }
}

public struct ScanReport: Codable {
    public var diskTotalBytes: Int64
    public var diskFreeBytes: Int64
    public var diskUsedBytes: Int64
    public var diskCapacityPercent: Int
    public var totalReclaimableBytes: Int64
    public var groups: [Group]

    enum CodingKeys: String, CodingKey {
        case diskTotalBytes = "disk_total_bytes"
        case diskFreeBytes = "disk_free_bytes"
        case diskUsedBytes = "disk_used_bytes"
        case diskCapacityPercent = "disk_capacity_percent"
        case totalReclaimableBytes = "total_reclaimable_bytes"
        case groups
    }

    public var totalReclaimableFormatted: String {
        ByteFormatter.format(totalReclaimableBytes)
    }
}

public struct CleanResult: Codable {
    public var freedBytes: Int64
    public var freedFormatted: String
    public var cleanedCount: Int
    public var dryRun: Bool?
    public var errors: [String]?

    enum CodingKeys: String, CodingKey {
        case freedBytes = "freed_bytes"
        case freedFormatted = "freed_formatted"
        case cleanedCount = "cleaned_count"
        case dryRun = "dry_run"
        case errors
    }
}

public enum ByteFormatter {
    public static func format(_ bytes: Int64) -> String {
        let b = Double(bytes)
        let kb = b / 1024
        let mb = kb / 1024
        let gb = mb / 1024
        let tb = gb / 1024

        if tb >= 1.0 {
            return String(format: "%.2f TB", tb)
        } else if gb >= 1.0 {
            return String(format: "%.2f GB", gb)
        } else if mb >= 1.0 {
            return String(format: "%.2f MB", mb)
        } else if kb >= 1.0 {
            return String(format: "%.2f KB", kb)
        } else {
            return "\(bytes) B"
        }
    }
}

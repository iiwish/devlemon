//go:build darwin

package probe

import (
	"encoding/binary"
	"path/filepath"
	"sync"
	"unsafe"

	"golang.org/x/sys/unix"
)

// getattrlistbulk 批量读取目录项属性：一次系统调用返回一批条目的名称、类型与大小，
// 省去逐文件 lstat，对 node_modules / target 这类海量小文件目录比 ReadDir+lstat 快 2~5 倍。
// 该调用只用于统计体积，绝不参与决定删除哪些路径；任何异常都回退到 ReadDir+lstat。

const (
	attrBitMapCount     = 5
	attrCmnName         = 0x00000001
	attrCmnObjType      = 0x00000008
	attrCmnReturnedAttr = 0x80000000
	attrFileDataLength  = 0x00000200
	fsoptPackInvalAttrs = 0x00000008
	vDir                = 2

	// 条目布局：length u32 | returned attribute_set_t (common, vol, dir, file, fork 各 u32) |
	// name attrreference_t (i32 相对偏移, u32 长度含 NUL) | objtype u32 | [datalength i64，仅当 returned.file 含该位]
	// 目录条目不含文件属性，因此必须依据内核实际返回的属性位图解析，不能假定固定长度
	entryReturnedFileOff = 4 + 4*3
	entryNameRefOff      = 24
	entryObjTypeOff      = 32
	entryFixedLen        = 36
	entryDataLenSize     = 8
)

type attrList struct {
	bitmapCount uint16
	reserved    uint16
	commonAttr  uint32
	volAttr     uint32
	dirAttr     uint32
	fileAttr    uint32
	forkAttr    uint32
}

var bulkBufPool = sync.Pool{New: func() any { b := make([]byte, 64*1024); return &b }}

// readDirSizes 返回目录下所有非目录条目的大小之和，以及子目录列表（不跟随符号链接）
func readDirSizes(dir string) (int64, []string) {
	if total, subdirs, ok := bulkReadDirSizes(dir); ok {
		return total, subdirs
	}
	return genericReadDirSizes(dir)
}

func bulkReadDirSizes(dir string) (int64, []string, bool) {
	fd, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
	if err != nil {
		return 0, nil, false
	}
	defer unix.Close(fd)

	bp := bulkBufPool.Get().(*[]byte)
	defer bulkBufPool.Put(bp)
	buf := *bp

	al := attrList{
		bitmapCount: attrBitMapCount,
		commonAttr:  attrCmnReturnedAttr | attrCmnName | attrCmnObjType,
		fileAttr:    attrFileDataLength,
	}

	var total int64
	var subdirs []string
	for {
		n, _, errno := unix.Syscall6(unix.SYS_GETATTRLISTBULK, uintptr(fd),
			uintptr(unsafe.Pointer(&al)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)), fsoptPackInvalAttrs, 0)
		if errno == unix.EINTR {
			continue
		}
		if errno != 0 {
			return 0, nil, false
		}
		if n == 0 {
			return total, subdirs, true
		}

		off := 0
		for i := 0; i < int(n); i++ {
			if off+entryFixedLen > len(buf) {
				return 0, nil, false
			}
			entry := buf[off:]
			length := int(binary.LittleEndian.Uint32(entry))
			if length < entryFixedLen || off+length > len(buf) {
				return 0, nil, false
			}
			entry = entry[:length]
			returnedFile := binary.LittleEndian.Uint32(entry[entryReturnedFileOff:])

			if binary.LittleEndian.Uint32(entry[entryObjTypeOff:]) == vDir {
				nameOff := int(int32(binary.LittleEndian.Uint32(entry[entryNameRefOff:])))
				nameLen := int(binary.LittleEndian.Uint32(entry[entryNameRefOff+4:]))
				start := entryNameRefOff + nameOff
				// nameLen 含结尾 NUL；名称位于固定字段之后
				if nameLen < 2 || start < entryFixedLen || start+nameLen > length {
					return 0, nil, false
				}
				subdirs = append(subdirs, filepath.Join(dir, string(entry[start:start+nameLen-1])))
			} else if returnedFile&attrFileDataLength != 0 {
				if length < entryFixedLen+entryDataLenSize {
					return 0, nil, false
				}
				total += int64(binary.LittleEndian.Uint64(entry[entryFixedLen:]))
			} else {
				// 文件条目未返回大小属性：交给通用实现，保证统计准确
				return 0, nil, false
			}
			off += length
		}
	}
}

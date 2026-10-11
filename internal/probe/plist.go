package probe

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"
)

// 轻量 plist 解码器，支持 XML 与二进制 (bplist00) 两种格式，解码为 Go 基础类型：
// dict -> map[string]any，array -> []any，string -> string，integer -> int64，
// real -> float64，true/false -> bool，data -> []byte，date -> string（原样保留）。
// 只用于读取 Info.plist、launchd 配置等小文件，不支持写入。

var errBadPlist = errors.New("invalid plist")

// readPlistFile 读取并解码 plist 文件
func readPlistFile(path string) (any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decodePlist(data)
}

func decodePlist(data []byte) (any, error) {
	if bytes.HasPrefix(data, []byte("bplist00")) {
		return decodeBinaryPlist(data)
	}
	return decodeXMLPlist(data)
}

// plistDictString 读取顶层 dict 中的字符串字段
func plistDictString(v any, key string) string {
	if m, ok := v.(map[string]any); ok {
		if s, ok := m[key].(string); ok {
			return s
		}
	}
	return ""
}

// ---------- XML ----------

func decodeXMLPlist(data []byte) (any, error) {
	d := xml.NewDecoder(bytes.NewReader(data))
	d.Strict = false
	for {
		tok, err := d.Token()
		if err != nil {
			return nil, errBadPlist
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local != "plist" {
			return xmlValue(d, se)
		}
	}
}

func xmlValue(d *xml.Decoder, se xml.StartElement) (any, error) {
	switch se.Name.Local {
	case "dict":
		m := make(map[string]any)
		var key string
		haveKey := false
		for {
			tok, err := d.Token()
			if err != nil {
				return nil, errBadPlist
			}
			switch t := tok.(type) {
			case xml.StartElement:
				if t.Name.Local == "key" {
					if key, err = xmlText(d); err != nil {
						return nil, err
					}
					haveKey = true
					continue
				}
				v, err := xmlValue(d, t)
				if err != nil {
					return nil, err
				}
				if haveKey {
					m[key] = v
					haveKey = false
				}
			case xml.EndElement:
				return m, nil
			}
		}
	case "array":
		var arr []any
		for {
			tok, err := d.Token()
			if err != nil {
				return nil, errBadPlist
			}
			switch t := tok.(type) {
			case xml.StartElement:
				v, err := xmlValue(d, t)
				if err != nil {
					return nil, err
				}
				arr = append(arr, v)
			case xml.EndElement:
				return arr, nil
			}
		}
	case "true", "false":
		if err := d.Skip(); err != nil {
			return nil, errBadPlist
		}
		return se.Name.Local == "true", nil
	}

	text, err := xmlText(d)
	if err != nil {
		return nil, err
	}
	switch se.Name.Local {
	case "integer":
		n, err := strconv.ParseInt(strings.TrimSpace(text), 10, 64)
		if err != nil {
			return nil, errBadPlist
		}
		return n, nil
	case "real":
		f, err := strconv.ParseFloat(strings.TrimSpace(text), 64)
		if err != nil {
			return nil, errBadPlist
		}
		return f, nil
	case "data":
		return []byte(strings.TrimSpace(text)), nil
	default: // string、date 及未知标签按字符串处理
		return text, nil
	}
}

// xmlText 读取当前元素的文本内容直到其结束标签
func xmlText(d *xml.Decoder) (string, error) {
	var sb strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			return "", errBadPlist
		}
		switch t := tok.(type) {
		case xml.CharData:
			sb.Write(t)
		case xml.EndElement:
			return sb.String(), nil
		case xml.StartElement:
			if err := d.Skip(); err != nil {
				return "", errBadPlist
			}
		}
	}
}

// ---------- 二进制 bplist00 ----------

type bplist struct {
	data       []byte
	offsets    []uint64
	refSize    int
	depthGuard int
	budget     int // 剩余可解码对象数，防止自引用对象导致指数级展开
}

func decodeBinaryPlist(data []byte) (any, error) {
	if len(data) < 8+32 {
		return nil, errBadPlist
	}
	trailer := data[len(data)-32:]
	offsetSize := int(trailer[6])
	refSize := int(trailer[7])
	numObjects := binary.BigEndian.Uint64(trailer[8:16])
	topObject := binary.BigEndian.Uint64(trailer[16:24])
	tableOffset := binary.BigEndian.Uint64(trailer[24:32])
	if offsetSize < 1 || offsetSize > 8 || refSize < 1 || refSize > 8 ||
		numObjects == 0 || topObject >= numObjects ||
		tableOffset >= uint64(len(data)) || numObjects > uint64(len(data))/uint64(offsetSize) {
		return nil, errBadPlist
	}
	p := &bplist{data: data, refSize: refSize, offsets: make([]uint64, numObjects), budget: 4*int(numObjects) + 64}
	for i := range p.offsets {
		start := tableOffset + uint64(i*offsetSize)
		if start+uint64(offsetSize) > uint64(len(data)) {
			return nil, errBadPlist
		}
		p.offsets[i] = beUint(data[start : start+uint64(offsetSize)])
	}
	return p.object(topObject)
}

func beUint(b []byte) uint64 {
	var n uint64
	for _, c := range b {
		n = n<<8 | uint64(c)
	}
	return n
}

// object 解码第 ref 个对象；嵌套深度与解码总量均受限，防止恶意或损坏文件导致无限递归或指数级展开
func (p *bplist) object(ref uint64) (any, error) {
	if ref >= uint64(len(p.offsets)) {
		return nil, errBadPlist
	}
	p.depthGuard++
	p.budget--
	defer func() { p.depthGuard-- }()
	if p.depthGuard > 64 || p.budget < 0 {
		return nil, errBadPlist
	}

	off := p.offsets[ref]
	if off >= uint64(len(p.data)) {
		return nil, errBadPlist
	}
	marker := p.data[off]
	kind, info := marker>>4, marker&0x0f
	pos := off + 1

	switch kind {
	case 0x0:
		switch marker {
		case 0x08:
			return false, nil
		case 0x09:
			return true, nil
		}
		return nil, nil
	case 0x1: // int：2^info 字节大端
		n := uint64(1) << info
		b, err := p.slice(pos, n)
		if err != nil {
			return nil, err
		}
		return int64(beUint(b)), nil
	case 0x2: // real
		n := uint64(1) << info
		b, err := p.slice(pos, n)
		if err != nil {
			return nil, err
		}
		switch n {
		case 4:
			return float64(math.Float32frombits(uint32(beUint(b)))), nil
		case 8:
			return math.Float64frombits(beUint(b)), nil
		}
		return nil, errBadPlist
	case 0x3: // date：不需要，保留占位
		return "", nil
	}

	count, pos, err := p.count(info, pos)
	if err != nil {
		return nil, err
	}
	switch kind {
	case 0x4: // data
		return p.slice(pos, count)
	case 0x5: // ASCII 字符串
		b, err := p.slice(pos, count)
		if err != nil {
			return nil, err
		}
		return string(b), nil
	case 0x6: // UTF-16BE 字符串，count 为字符数
		b, err := p.slice(pos, count*2)
		if err != nil {
			return nil, err
		}
		u := make([]uint16, count)
		for i := range u {
			u[i] = binary.BigEndian.Uint16(b[i*2:])
		}
		return string(utf16.Decode(u)), nil
	case 0xA: // array
		refs, err := p.refs(pos, count)
		if err != nil {
			return nil, err
		}
		arr := make([]any, 0, len(refs))
		for _, r := range refs {
			v, err := p.object(r)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil
	case 0xD: // dict：先 count 个 key 引用，再 count 个 value 引用
		refs, err := p.refs(pos, count*2)
		if err != nil {
			return nil, err
		}
		m := make(map[string]any, count)
		for i := uint64(0); i < count; i++ {
			k, err := p.object(refs[i])
			if err != nil {
				return nil, err
			}
			ks, ok := k.(string)
			if !ok {
				return nil, errBadPlist
			}
			v, err := p.object(refs[count+i])
			if err != nil {
				return nil, err
			}
			m[ks] = v
		}
		return m, nil
	}
	// UID、set 等本场景用不到的类型
	return nil, nil
}

// count 读取对象长度：低 4 位为 0xF 时长度存放在紧随其后的 int 对象中
func (p *bplist) count(info byte, pos uint64) (uint64, uint64, error) {
	if info != 0x0f {
		return uint64(info), pos, nil
	}
	if pos >= uint64(len(p.data)) || p.data[pos]>>4 != 0x1 {
		return 0, 0, errBadPlist
	}
	n := uint64(1) << (p.data[pos] & 0x0f)
	b, err := p.slice(pos+1, n)
	if err != nil {
		return 0, 0, err
	}
	return beUint(b), pos + 1 + n, nil
}

func (p *bplist) refs(pos, n uint64) ([]uint64, error) {
	b, err := p.slice(pos, n*uint64(p.refSize))
	if err != nil {
		return nil, err
	}
	refs := make([]uint64, n)
	for i := range refs {
		refs[i] = beUint(b[i*p.refSize : (i+1)*p.refSize])
	}
	return refs, nil
}

func (p *bplist) slice(pos, n uint64) ([]byte, error) {
	if n > uint64(len(p.data)) || pos > uint64(len(p.data))-n {
		return nil, errBadPlist
	}
	return p.data[pos : pos+n], nil
}

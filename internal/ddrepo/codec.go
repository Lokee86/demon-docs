package ddrepo

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	shardMagic      = "DDOC"
	shardVersion    = byte(1)
	shardHeaderSize = len(shardMagic) + 1 + 4
	shardLengthSize = 4
)

func shardName(name string) string {
	digest := sha256.Sum256([]byte(name))
	return hex.EncodeToString(digest[:1])[:1]
}

func encodeShard(records map[string][]byte) ([]byte, error) {
	names := make([]string, 0, len(records))
	for name, value := range records {
		if err := validateRecordName(name); err != nil {
			return nil, err
		}
		if uint64(len(name)) > uint64(^uint32(0)) {
			return nil, fmt.Errorf("record name is too large: %q", name)
		}
		if uint64(len(value)) > uint64(^uint32(0)) {
			return nil, fmt.Errorf("record value is too large: %q", name)
		}
		names = append(names, name)
	}
	sort.Strings(names)

	size := shardHeaderSize
	for _, name := range names {
		size += shardLengthSize + shardLengthSize + len(name) + len(records[name])
	}
	encoded := make([]byte, size)
	copy(encoded, shardMagic)
	encoded[len(shardMagic)] = shardVersion
	binary.BigEndian.PutUint32(encoded[len(shardMagic)+1:], uint32(len(names)))

	offset := shardHeaderSize
	for _, name := range names {
		value := records[name]
		binary.BigEndian.PutUint32(encoded[offset:], uint32(len(name)))
		offset += shardLengthSize
		binary.BigEndian.PutUint32(encoded[offset:], uint32(len(value)))
		offset += shardLengthSize
		copy(encoded[offset:], name)
		offset += len(name)
		copy(encoded[offset:], value)
		offset += len(value)
	}
	return encoded, nil
}

func decodeShard(data []byte) (map[string][]byte, error) {
	if len(data) < shardHeaderSize {
		return nil, fmt.Errorf("malformed shard: header is truncated")
	}
	if string(data[:len(shardMagic)]) != shardMagic {
		return nil, fmt.Errorf("malformed shard: invalid magic header")
	}
	if data[len(shardMagic)] != shardVersion {
		return nil, fmt.Errorf("unsupported shard version: %d", data[len(shardMagic)])
	}

	offset := len(shardMagic) + 1
	count := binary.BigEndian.Uint32(data[offset:])
	offset += shardLengthSize
	records := make(map[string][]byte)
	for index := uint32(0); index < count; index++ {
		nameLength, err := readShardLength(data, &offset, "name")
		if err != nil {
			return nil, err
		}
		valueLength, err := readShardLength(data, &offset, "value")
		if err != nil {
			return nil, err
		}
		nameBytes, err := readShardField(data, &offset, nameLength, "name")
		if err != nil {
			return nil, err
		}
		value, err := readShardField(data, &offset, valueLength, "value")
		if err != nil {
			return nil, err
		}
		name := string(nameBytes)
		if err := validateRecordName(name); err != nil {
			return nil, err
		}
		if _, exists := records[name]; exists {
			return nil, fmt.Errorf("malformed shard: duplicate record %q", name)
		}
		records[name] = append([]byte(nil), value...)
	}
	if offset != len(data) {
		return nil, fmt.Errorf("malformed shard: trailing data")
	}
	return records, nil
}

func validateRecordName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid record name: empty name")
	}
	if !utf8.ValidString(name) {
		return fmt.Errorf("invalid record name %q: invalid UTF-8", name)
	}
	if strings.HasPrefix(name, "/") || isDriveAbsoluteName(name) {
		return fmt.Errorf("invalid record name %q: absolute name", name)
	}
	if strings.ContainsRune(name, '\\') {
		return fmt.Errorf("invalid record name %q: backslash is not allowed", name)
	}
	if strings.IndexByte(name, 0) >= 0 {
		return fmt.Errorf("invalid record name %q: NUL is not allowed", name)
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "" {
			return fmt.Errorf("invalid record name %q: empty segment", name)
		}
		if segment == "." || segment == ".." {
			return fmt.Errorf("invalid record name %q: invalid segment %q", name, segment)
		}
	}
	return nil
}

func isDriveAbsoluteName(name string) bool {
	return len(name) >= 3 &&
		((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z')) &&
		name[1] == ':' && name[2] == '/'
}

func readShardLength(data []byte, offset *int, field string) (uint32, error) {
	if len(data)-*offset < shardLengthSize {
		return 0, fmt.Errorf("malformed shard: truncated %s length", field)
	}
	length := binary.BigEndian.Uint32(data[*offset:])
	*offset += shardLengthSize
	return length, nil
}

func readShardField(data []byte, offset *int, length uint32, field string) ([]byte, error) {
	remaining := len(data) - *offset
	if uint64(length) > uint64(remaining) {
		return nil, fmt.Errorf("malformed shard: truncated %s data", field)
	}
	end := *offset + int(length)
	fieldData := data[*offset:end]
	*offset = end
	return fieldData, nil
}

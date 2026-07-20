package frontmatter

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"regexp"
	"time"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func validateValue(kind string, value any) (any, error) {
	kind = normalizedType(kind)
	value = normalize(value)
	switch kind {
	case "string":
		if _, ok := value.(string); !ok {
			return value, fmt.Errorf("expected string, got %T", value)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return value, fmt.Errorf("expected boolean, got %T", value)
		}
	case "integer":
		if _, ok := value.(int64); !ok {
			return value, fmt.Errorf("expected integer, got %T", value)
		}
	case "number":
		switch value.(type) {
		case int64, float64:
		default:
			return value, fmt.Errorf("expected number, got %T", value)
		}
	case "string_list":
		list, ok := value.([]any)
		if !ok {
			return value, fmt.Errorf("expected string list, got %T", value)
		}
		for _, item := range list {
			if _, ok := item.(string); !ok {
				return value, fmt.Errorf("expected string list, got item %T", item)
			}
		}
	case "date":
		text, ok := value.(string)
		if !ok {
			if stringer, ok := value.(fmt.Stringer); ok {
				text = stringer.String()
			} else {
				return value, fmt.Errorf("expected date, got %T", value)
			}
		}
		parsed, err := time.Parse("2006-01-02", text)
		if err != nil {
			return value, fmt.Errorf("expected date in YYYY-MM-DD form")
		}
		return parsed.Format("2006-01-02"), nil
	case "uuid":
		text, ok := value.(string)
		if !ok || !uuidPattern.MatchString(text) {
			return value, fmt.Errorf("expected UUID")
		}
	default:
		return value, fmt.Errorf("unsupported configured field type %q", kind)
	}
	return value, nil
}

func NewUUIDv7(now time.Time) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	milliseconds := uint64(now.UnixMilli())
	raw[0], raw[1], raw[2] = byte(milliseconds>>40), byte(milliseconds>>32), byte(milliseconds>>24)
	raw[3], raw[4], raw[5] = byte(milliseconds>>16), byte(milliseconds>>8), byte(milliseconds)
	raw[6] = (raw[6] & 0x0f) | 0x70
	raw[8] = (raw[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(raw[0:4]),
		binary.BigEndian.Uint16(raw[4:6]),
		binary.BigEndian.Uint16(raw[6:8]),
		binary.BigEndian.Uint16(raw[8:10]),
		uint64(raw[10])<<40|uint64(raw[11])<<32|uint64(raw[12])<<24|uint64(raw[13])<<16|uint64(raw[14])<<8|uint64(raw[15])), nil
}

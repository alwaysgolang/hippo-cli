package plugins

import (
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

func GetJSONKeyInsensitive(m map[string]gjson.Result, key string) (*gjson.Result, bool) {
	keyParts := strings.Split(key, ".")
	partsLength := len(keyParts)

	for i, part := range keyParts {
		isReturnPart := partsLength == i+1

		if isReturnPart {
			for jsonKey, value := range m {
				if strings.EqualFold(jsonKey, part) && value.Exists() && value.Type != gjson.Null {
					return &value, true
				}
			}
		} else {
			for jsonKey, value := range m {
				if strings.EqualFold(jsonKey, part) {
					m = value.Map()
					break
				}
			}
		}
	}
	return nil, false
}

func SafeGet[T any](m map[string]gjson.Result, keys ...string) T {
	var zero T
	if m == nil {
		return zero
	}
	var val *gjson.Result
	var found bool
	for _, key := range keys {
		val, found = GetJSONKeyInsensitive(m, key)
		if found {
			break
		}
	}

	if !found {
		switch any(zero).(type) {
		case *string, *int, *int16, *int64, *float64, *bool:
			var nilPtr T
			return nilPtr
		case *DateFormat:
			zeroFormat := DateFormat(time.Time{})
			return any(&zeroFormat).(T)
		case *DateTimeFormat:
			zeroFormat := DateTimeFormat(time.Time{})
			return any(&zeroFormat).(T)
		case DateFormat:
			now := DateFormat(GetNow())
			return any(now).(T)
		case DateTimeFormat:
			now := DateTimeFormat(GetNow())
			return any(now).(T)
		default:
			return zero
		}
	}

	if result, ok := any(val.Map()).(T); ok {
		return result
	}

	strVal := val.String()
	return genericParse[T](strVal, ",")
}

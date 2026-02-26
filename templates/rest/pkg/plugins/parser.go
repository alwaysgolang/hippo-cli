package plugins

import (
	"strconv"
	"strings"
	"time"
)

func genericParse[T any](query, splitKey string) T {
	var zero T
	if query == "" {
		switch any(zero).(type) {
		case *DateFormat:
			d := DateFormat(time.Time{})
			return any(&d).(T)
		case *DateTimeFormat:
			d := DateTimeFormat(time.Time{})
			return any(&d).(T)
		case DateFormat, DateTimeFormat:
		default:
			return zero
		}
	}

	switch any(zero).(type) {
	case string:
		return any(query).(T)
	case int:
		i, _ := strconv.Atoi(query)
		return any(i).(T)
	case int16:
		i, _ := strconv.Atoi(query)
		return any(int16(i)).(T)
	case int64:
		i, _ := strconv.ParseInt(query, 10, 64)
		return any(i).(T)
	case float64:
		f, _ := strconv.ParseFloat(query, 64)
		return any(f).(T)
	case bool:
		b, _ := strconv.ParseBool(query)
		return any(b).(T)

	case *string:
		return any(&query).(T)
	case *int:
		i, _ := strconv.Atoi(query)
		return any(&i).(T)
	case *int16:
		i, _ := strconv.Atoi(query)
		i16 := int16(i)
		return any(&i16).(T)
	case *int64:
		i, _ := strconv.ParseInt(query, 10, 64)
		i64 := i
		return any(&i64).(T)
	case *bool:
		b, _ := strconv.ParseBool(query)
		return any(&b).(T)
	case DateFormat:
		d := parseDateFormat(query)
		return any(d).(T)
	case *DateFormat:
		d := parseDateFormat(query)
		return any(&d).(T)
	case DateTimeFormat:
		d := parseDateTimeFormat(query)
		return any(d).(T)
	case *DateTimeFormat:
		d := parseDateTimeFormat(query)
		return any(&d).(T)

	case []int:
		return any(genericParseSlice[int](query, splitKey)).(T)
	case []int16:
		return any(genericParseSlice[int16](query, splitKey)).(T)
	case []int64:
		return any(genericParseSlice[int64](query, splitKey)).(T)
	case []string:
		return any(genericParseSlice[string](query, splitKey)).(T)
	case []float64:
		return any(genericParseSlice[float64](query, splitKey)).(T)
	case []bool:
		return any(genericParseSlice[bool](query, splitKey)).(T)

	default:
		return zero
	}
}

func parseDateFormat(query string) DateFormat {
	if t, err := time.ParseInLocation(time.RFC3339, query, location); err == nil {
		return DateFormat(t)
	}
	if t, err := time.ParseInLocation(time.DateOnly, query, location); err == nil {
		return DateFormat(t)
	}
	return DateFormat(GetNow())
}

func parseDateTimeFormat(query string) DateTimeFormat {
	if t, err := time.ParseInLocation(time.RFC3339, query, location); err == nil {
		return DateTimeFormat(t)
	}
	if t, err := time.ParseInLocation(time.DateTime, query, location); err == nil {
		return DateTimeFormat(t)
	}
	return DateTimeFormat(GetNow())
}

func genericParseSlice[T any](query, splitKey string) []T {
	items := strings.Split(query, ",")
	result := make([]T, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		result = append(result, genericParse[T](item, splitKey))
	}
	return result
}

func ParseQuery[T any](query string) T {
	return genericParse[T](query, ",")
}

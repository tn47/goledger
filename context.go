package main

import "fmt"

// Context map of context values
type Context map[string]interface{}

func (context Context) Int64(key string) (int64, bool) {
	value, ok := context[key]
	if !ok {
		return 0, ok
	}
	switch val := value.(type) {
	case float64:
		return int64(val), true
	case float32:
		return int64(val), true
	case uint:
		return int64(val), true
	case uint64:
		return int64(val), true
	case uint32:
		return int64(val), true
	case uint16:
		return int64(val), true
	case uint8:
		return int64(val), true
	case int:
		return int64(val), true
	case int64:
		return int64(val), true
	case int32:
		return int64(val), true
	case int16:
		return int64(val), true
	case int8:
		return int64(val), true
	}
	panic(fmt.Errorf("context %v not a number: %T", key, value))
}

// String return the string value for key.
func (context Context) String(key string) (string, bool) {
	value, ok := context[key]
	if !ok {
		return "", ok
	}
	val, ok := value.(string)
	return val, ok
}

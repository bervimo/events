package utils

import (
	"errors"
	"os"
	"strconv"
)

// Value
type Value interface {
	string | int | int8 | int16 | int32 | int64
}

// ErrNotFound
var ErrNotFound = errors.New("utils: env var not found")

// GetEnv
func GetEnv[V Value](key string) (V, error) {
	val, exists := os.LookupEnv(key)

	var value V

	switch any(value).(type) {
	case string:
		if !exists {
			return any(val).(V), ErrNotFound
		}

		return any(val).(V), nil
	case int:
		if !exists {
			return any(0).(V), ErrNotFound
		}

		iVal, err := strconv.Atoi(val)

		return any(iVal).(V), err
	}

	return any(nil).(V), nil
}

// GetEnvOr
func GetEnvOr[V Value](key string, def V) (V, error) {
	val, exists := os.LookupEnv(key)

	var value V

	switch any(value).(type) {
	case string:
		if !exists {
			return any(def).(V), ErrNotFound
		}

		return any(val).(V), nil
	case int:
		if !exists {
			return any(def).(V), ErrNotFound
		}

		iVal, err := strconv.Atoi(val)

		return any(iVal).(V), err
	}

	return any(nil).(V), nil
}

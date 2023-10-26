package oss

import "time"

// Bool returns a pointer value for the bool value.
func Bool(v bool) *bool {
	return &v
}

// String returns a pointer value for the string value.
func String(v string) *string {
	return &v
}

// StringSlice returns a slice of string pointers from the values.
func StringSlice(vs []string) []*string {
	ps := make([]*string, len(vs))
	for i, v := range vs {
		vv := v
		ps[i] = &vv
	}

	return ps
}

// StringMap returns a map of string pointers from the values.
func StringMap(vs map[string]string) map[string]*string {
	ps := make(map[string]*string, len(vs))
	for k, v := range vs {
		vv := v
		ps[k] = &vv
	}

	return ps
}

// Int returns a pointer value for the int.
func Int(v int) *int {
	return &v
}

// Int64 returns a pointer value for the int64.
func Int64(v int64) *int64 {
	return &v
}

// Time returns a pointer value for the time.
func Time(v time.Time) *time.Time {
	return &v
}

// Duration returns a pointer value for the time.
func Duration(v time.Duration) *time.Duration {
	return &v
}

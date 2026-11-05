package utils

import "time"

func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func SafeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

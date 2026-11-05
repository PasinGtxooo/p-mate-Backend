package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
)

// ShortUUIDWithPrefix convert UUID to short UUID with prefix
func ShortUUIDWithPrefix(id uuid.UUID, prefix string) string {
	shortID := shortuuid.DefaultEncoder.Encode(id)
	return fmt.Sprintf("%s_%s", prefix, shortID)
}

// HasIDPrefix checks if the ID starts with the given prefix
func HasIDPrefix(id string, prefix string) bool {
	return strings.HasPrefix(id, prefix+"_")
}

// ParseID parses an ID that could be either a short UUID with prefix or a standard UUID
func ParseID(id string, prefix string) (uuid.UUID, error) {
	if HasIDPrefix(id, prefix) {
		prefixLen := len(prefix) + 1
		if len(id) < prefixLen {
			return uuid.Nil, fmt.Errorf("invalid short UUID format")
		}
		actualShortID := id[prefixLen:]
		return shortuuid.DefaultEncoder.Decode(actualShortID)
	}
	return uuid.Parse(id)
}

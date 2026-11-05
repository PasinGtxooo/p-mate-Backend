package entity

import "time"

// File represents a file stored in Storage
type File struct {
	Key       string
	Name      string
	Size      int64
	MimeType  string
	Extension string
	CreatedAt time.Time
}

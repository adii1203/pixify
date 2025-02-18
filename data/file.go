package data

import (
	"gorm.io/gorm"
)

type File struct {
	gorm.Model
	Path     string `json:"path"`
	Name     string `json:"name"`
	FileType string `json:"file_type"`
	MimeType string `json:"mime_type"`
	Size     uint64 `json:"size"`
	Width    uint64 `json:"width"`
	Height   uint64 `json:"height"`
	Url      string `json:"url"`
}

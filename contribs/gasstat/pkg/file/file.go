package file

import (
	"path/filepath"
	"strings"
)

// FileFormat represents an enumeration of file formats.
type FileFormat uint8

// Enumeration of file formats.
const (
	Unknown FileFormat = iota
	JSON
	YAML
	CSV
	TXT
	MARKDOWN
)

// String implements the Stringer interface for FileFormat.
func (f FileFormat) String() string {
	return []string{"Unknown", "JSON", "YAML", "CSV", "TXT", "MARKDOWN"}[f]
}

// FileFormatFromExt returns the FileFormat based on the file extension.
func FileFormatFromExt(filename string) FileFormat {
	// Extract the file extension and convert it to lowercase.
	extension := filepath.Ext(filename)
	extension = strings.ToLower(extension)

	// Determine the file format based on the extension.
	switch extension {
	case ".json":
		return JSON
	case ".yaml", ".yml":
		return YAML
	case ".csv":
		return CSV
	case ".txt":
		return TXT
	case ".md":
		return MARKDOWN
	default:
		return Unknown
	}
}

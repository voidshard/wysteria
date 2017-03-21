package searchends

import (
	"strings"
	"encoding/base64"
)

// util to avoid odd chars or tokenizing via b64 encoding the string
func b64encode(path string) string {
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(path)), "=")
}

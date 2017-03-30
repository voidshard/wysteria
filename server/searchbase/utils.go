package searchends

import (
	"encoding/base64"
	"strings"
)

// util to avoid odd chars or tokenizing via b64 encoding the string
func b64encode(path string) string {
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(path)), "=")
}

package searchends

const (
	// general names for indexes - provided simply for some consistency
	// and to avoid having random strings used about the place.
	tableCollection = "collections"
	tableItem       = "items"
	tableVersion    = "versions"
	tableResource   = "fileresource"
	tableLink       = "link"
)

// Settings that will be provided to a searchbase implementation on creation
type Settings struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Pass     string
	Database string
	PemFile  string
}

package searchends

const (
	tableCollection   = "collections"
	tableItem         = "items"
	tableVersion      = "versions"
	tableFileresource = "fileresource"
	tableLink         = "link"
)

type SearchbaseSettings struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Pass     string
	Database string
	PemFile  string
}

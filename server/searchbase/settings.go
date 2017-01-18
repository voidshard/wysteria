package searchends

const (
	table_item         = "items"
	table_version      = "versions"
	table_fileresource = "fileresource"
	table_link         = "link"
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

package database

const (
	table_collection = "collections"
	table_item = "items"
	table_version = "versions"
	table_fileresource = "fileresource"
	table_link = "link"
)

type DatabaseSettings struct {
	Driver string
	Host string
	Port int
	User string
	Pass string
	Database string
	PemFile string
}

package database

const (
	// Default names for the tables / collections / indexes / buckets
	// The use of these isn't required, they're just provided in the db module for
	// some consistency in naming.
	tableCollection   = "collections"
	tableItem         = "items"
	tableVersion      = "versions"
	tableFileresource = "fileresource"
	tableLink         = "link"
)

type Settings struct {
	// The name of the driver (will be used to load via Connect in database.go)
	Driver   string

	// Database settings to use in the connection
	Host     string
	Port     int
	User     string
	Pass     string
	Database string
	PemFile  string
}

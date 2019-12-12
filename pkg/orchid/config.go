package orchid

// Config represents the primary configuration structure for Orchid.
type Config struct {
	PgDatabase string // postgresql database name
	PgUsername string // postgresql username
	PgPassword string // postgresql password
	PgConnStr  string // libpq extra connection string
}

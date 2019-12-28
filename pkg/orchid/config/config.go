package config

// Config represents the primary configuration structure for Orchid.
type Config struct {
	Username string // postgresql username
	Password string // postgresql password
	Options  string // key=value set of libpq connection string options
}

package postgres

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresConfig holds PostgreSQL-specific configuration
type PostgresConfig struct {
	Host     string     `koanf:"host" json:"host"`
	Port     int        `koanf:"port" json:"port"`
	Database string     `koanf:"database" json:"database"`
	Username string     `koanf:"username" json:"username"`
	Password string     `koanf:"password" json:"password"`
	SSLMode  string     `koanf:"ssl_mode" json:"ssl_mode"`
	Pool     PoolConfig `koanf:"pool" json:"pool"`
}

// PoolConfig holds database connection pool configuration
type PoolConfig struct {
	MaxOpenConnections    int           `koanf:"max_open_connections" json:"max_open_connections"`
	MaxIdleConnections    int           `koanf:"max_idle_connections" json:"max_idle_connections"`
	ConnectionMaxLifetime time.Duration `koanf:"connection_max_lifetime" json:"connection_max_lifetime"`
}

// DatabaseConfig holds the configuration for database connections
type DatabaseConfig struct {
	Postgres PostgresConfig `koanf:"postgres" json:"postgres"`
}

// Config represents the complete application configuration
type Config struct {
	Database DatabaseConfig `koanf:"database" json:"database"`
}

// GetPostgresDSN builds a PostgreSQL connection string from the configuration
func (c *PostgresConfig) GetPostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.Username, c.Password, c.SSLMode)
}

// PostgresDAO implements the DAO interface using PostgreSQL and sqlx
type PostgresDAO struct {
	db *sqlx.DB
}

// NewPostgresDAO creates a new PostgreSQL DAO instance
func NewPostgresDAO(config *PostgresConfig) (*PostgresDAO, error) {
	dsn := config.GetPostgresDSN()

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Pool.MaxOpenConnections)
	db.SetMaxIdleConns(config.Pool.MaxIdleConnections)
	db.SetConnMaxLifetime(config.Pool.ConnectionMaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	slog.Info("PostgreSQL DAO created successfully",
		"host", config.Host,
		"port", config.Port,
		"database", config.Database,
		"max_open_conns", config.Pool.MaxOpenConnections)

	return &PostgresDAO{db: db}, nil
}

// NewPostgresDAOWithDB creates a new PostgreSQL DAO instance using a shared database connection
func NewPostgresDAOWithDB(db *sqlx.DB) (*PostgresDAO, error) {
	return &PostgresDAO{db: db}, nil
}

// Close closes the database connection
// Note: When using shared connection pool, this method should not be called
// as the connection is managed by the factory
func (p *PostgresDAO) Close() error {
	// Do nothing - connection is managed by the factory
	return nil
}

// Meter represents a meter record in the database
type Meter struct {
	ID             string `db:"id" json:"id"`
	IPv6           string `db:"ipv6" json:"ipv6"`
	Port           int    `db:"port" json:"port"`
	SystemTitle    string `db:"system_title" json:"system_title"`
	AuthPassword   string `db:"auth_password" json:"auth_password"`
	AuthKey        string `db:"auth_key" json:"auth_key"`
	BlockCipherKey string `db:"block_cipher_key" json:"block_cipher_key"`
}

// InsertMeter inserts a new meter record into the database
func (p *PostgresDAO) InsertMeter(meter *Meter) error {
	query := `
		INSERT INTO meters (id, ipv6, port, system_title, auth_password, auth_key, block_cipher_key)
		VALUES (:id, :ipv6, :port, :system_title, :auth_password, :auth_key, :block_cipher_key)
	`

	_, err := p.db.NamedExec(query, meter)
	if err != nil {
		return fmt.Errorf("failed to insert meter: %w", err)
	}

	slog.Info("Meter inserted successfully", "id", meter.ID)
	return nil
}

// GetMeter retrieves a meter record by ID
func (p *PostgresDAO) GetMeter(id string) (*Meter, error) {
	var meter Meter
	query := `SELECT id, ipv6, port, system_title, auth_password, auth_key, block_cipher_key FROM meters WHERE id = $1`

	err := p.db.Get(&meter, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get meter with id %s: %w", id, err)
	}

	return &meter, nil
}

// GetAllMeters retrieves all meter records
func (p *PostgresDAO) GetAllMeters() ([]Meter, error) {
	var meters []Meter
	query := `SELECT id, ipv6, port, system_title, auth_password, auth_key, block_cipher_key FROM meters`

	err := p.db.Select(&meters, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all meters: %w", err)
	}

	return meters, nil
}

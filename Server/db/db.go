// Package db provides database access for the OwnCord server.
// It uses modernc.org/sqlite — a pure-Go SQLite driver requiring no CGO.
package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/owncord/server/config"
	"github.com/owncord/server/migrations"
)

// DB wraps *sql.DB and exposes the subset of methods needed by the server.
type DB struct {
	sqlDB *sql.DB
}

// Open opens (or creates) a Postgresql database, and returns a ready-to-use DB.
func Open(path string) (*DB, error) {
	var connectString string
	cfg, err := config.Load("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	connectString = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.Database)

	sqlDB, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite db: %w", err)
	}

	// Verify the connection is actually usable.
	if err := sqlDB.Ping(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("pinging sqlite db: %w", err)
	}

	// SQLite only allows one writer at a time. Pin to a single connection
	// so concurrent goroutines queue on the Go side rather than getting
	// SQLITE_BUSY. For :memory: databases this also ensures all callers
	// share the same in-memory state.
	sqlDB.SetMaxOpenConns(1)

	return &DB{sqlDB: sqlDB}, nil
}

// Migrate runs all SQL migration files from the embedded migrations FS in
// lexicographic order, applying each file exactly once.  It delegates to
// MigrateFS (defined in migrate.go) which maintains the schema_versions
// tracking table.
func Migrate(database *DB) error {
	return MigrateFS(database, migrations.FS)
}

// Close releases the underlying database connection.
func (d *DB) Close() error {
	// Run optimize to analyze and update query planner statistics.
	_, _ = d.sqlDB.Exec("VACUUM FULL ANALYZE;")
	return d.sqlDB.Close()
}

// QueryRow executes a query that returns at most one row.
func (d *DB) QueryRow(query string, args ...any) *sql.Row {
	return d.sqlDB.QueryRow(query, args...)
}

// QueryRowContext executes a query that returns at most one row, with context.
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.sqlDB.QueryRowContext(ctx, query, args...)
}

// Exec executes a query that doesn't return rows.
func (d *DB) Exec(query string, args ...any) (sql.Result, error) {
	return d.sqlDB.Exec(query, args...)
}

// ExecContext executes a query that doesn't return rows, with context.
func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.sqlDB.ExecContext(ctx, query, args...)
}

// Query executes a query that returns multiple rows.
func (d *DB) Query(query string, args ...any) (*sql.Rows, error) {
	return d.sqlDB.Query(query, args...)
}

// QueryContext executes a query that returns multiple rows, with context.
func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.sqlDB.QueryContext(ctx, query, args...)
}

// Begin starts a database transaction.
func (d *DB) Begin() (*sql.Tx, error) {
	return d.sqlDB.Begin()
}

// BeginTx starts a database transaction with context and options.
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return d.sqlDB.BeginTx(ctx, opts)
}

// SQLDb returns the underlying *sql.DB for cases requiring direct access.
func (d *DB) SQLDb() *sql.DB {
	return d.sqlDB
}

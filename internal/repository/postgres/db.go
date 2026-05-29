package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/lib/pq"

	"github.com/Revvfi/revvfi-backend/internal/config"
	appErr "github.com/Revvfi/revvfi-backend/internal/pkg/errors"
)

/*
@struct DB

@desc
PostgreSQL database wrapper around database/sql.

@responsibilities
- Own SQL connection pool
- Provide transaction helper
- Expose repository constructors
*/
type DB struct {
	conn *sql.DB
}

/*
@function Open

@desc
Opens and verifies a PostgreSQL database connection.

@responsibilities
- Open database/sql connection
- Configure connection pool
- Ping database with context

@params
- ctx: startup lifecycle context
- cfg: database configuration

@returns
- *DB
- error
*/
func Open(ctx context.Context, cfg config.DatabaseConfig) (*DB, error) {
	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	conn.SetMaxOpenConns(cfg.MaxOpenConns)
	conn.SetMaxIdleConns(cfg.MaxIdleConns)
	conn.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := conn.PingContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &DB{conn: conn}, nil
}

/*
@function Wrap

@desc
Wraps an existing sql.DB for tests or externally managed pools.

@responsibilities
- Provide repository methods for an existing connection

@params
- conn: SQL database connection

@returns
- *DB
*/
func Wrap(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

/*
@method SQL

@desc
Returns the underlying database/sql connection.

@responsibilities
- Allow health checks and external integrations to use PingContext

@returns
- *sql.DB
*/
func (db *DB) SQL() *sql.DB {
	return db.conn
}

/*
@method Close

@desc
Closes the PostgreSQL connection pool.

@responsibilities
- Release database resources during shutdown

@returns
- error
*/
func (db *DB) Close() error {
	if db == nil || db.conn == nil {
		return nil
	}
	return db.conn.Close()
}

/*
@method PingContext

@desc
Checks database connectivity.

@responsibilities
- Delegate health checks to database/sql

@params
- ctx: request lifecycle context

@returns
- error
*/
func (db *DB) PingContext(ctx context.Context) error {
	if db == nil || db.conn == nil {
		return appErr.ErrDatabaseError
	}
	return db.conn.PingContext(ctx)
}

/*
@method WithTx

@desc
Runs a function inside a PostgreSQL transaction.

@responsibilities
- Begin transaction with context
- Roll back on function failure
- Commit on success

@params
- ctx: request lifecycle context
- fn: transactional callback

@returns
- error
*/
func (db *DB) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.conn.BeginTx(ctx, nil)
	if err != nil {
		return mapError(err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return mapError(err)
	}
	return nil
}

/*
@function mapError

@desc
Maps PostgreSQL and database/sql errors into domain errors.

@responsibilities
- Preserve not-found semantics
- Map unique constraint failures
- Wrap database failures consistently

@params
- err: source error

@returns
- error
*/
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return appErr.ErrDuplicateEntry
	}
	return fmt.Errorf("%w: %v", appErr.ErrDatabaseError, err)
}

/*
@function parseInt

@desc
Parses a decimal database string into big.Int.

@responsibilities
- Convert numeric text to big.Int
- Default empty values to zero

@params
- value: database string value

@returns
- *big.Int
*/
func parseInt(value string) *big.Int {
	if value == "" {
		return big.NewInt(0)
	}
	parsed, ok := new(big.Int).SetString(value, 10)
	if !ok {
		return big.NewInt(0)
	}
	return parsed
}

/*
@function now

@desc
Returns current UTC time for repository timestamps.

@responsibilities
- Keep timestamp generation consistent

@returns
- time.Time
*/
func now() time.Time {
	return time.Now().UTC()
}

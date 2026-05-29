package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/Revvfi/revvfi-backend/internal/models"
)

/*
@struct AuthRepository

@desc
PostgreSQL implementation for authentication persistence.

@responsibilities
- Store and validate SIWE nonces
- Store and revoke JWT sessions
- Check token revocation state
*/
type AuthRepository struct {
	db *DB
}

/*
@function NewAuthRepository

@desc
Creates a PostgreSQL auth repository.

@params
- db: PostgreSQL wrapper

@returns
- *AuthRepository
*/
func NewAuthRepository(db *DB) *AuthRepository {
	return &AuthRepository{db: db}
}

/*
@method StoreNonce

@desc
Stores an authentication nonce with expiration.
*/
func (r *AuthRepository) StoreNonce(ctx context.Context, wallet, nonce string, expiresAt time.Time) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into auth_nonces(wallet_address, nonce, expires_at, created_at) values($1,$2,$3,$4) on conflict(wallet_address) do update set nonce=$2, expires_at=$3, created_at=$4`, wallet, nonce, expiresAt, now())
	return mapError(err)
}

/*
@method ValidateNonce

@desc
Validates a stored authentication nonce.
*/
func (r *AuthRepository) ValidateNonce(ctx context.Context, wallet, nonce string) error {
	var stored string
	err := r.db.conn.QueryRowContext(ctx, `select nonce from auth_nonces where wallet_address=$1 and nonce=$2 and expires_at > now()`, wallet, nonce).Scan(&stored)
	return mapError(err)
}

/*
@method StoreSession

@desc
Persists an authenticated wallet session.
*/
func (r *AuthRepository) StoreSession(ctx context.Context, session *models.AuthSession) error {
	_, err := r.db.conn.ExecContext(ctx, `insert into auth_sessions(wallet_address, token, expires_at, created_at) values($1,$2,$3,$4)`, session.WalletAddress, session.Token, session.ExpiresAt, session.CreatedAt)
	return mapError(err)
}

/*
@method RevokeSession

@desc
Marks a JWT session as revoked.
*/
func (r *AuthRepository) RevokeSession(ctx context.Context, token string) error {
	_, err := r.db.conn.ExecContext(ctx, `update auth_sessions set revoked_at=$1 where token=$2`, now(), token)
	return mapError(err)
}

/*
@method IsSessionRevoked

@desc
Checks whether a JWT session has been revoked or expired.
*/
func (r *AuthRepository) IsSessionRevoked(ctx context.Context, token string) (bool, error) {
	var revokedAt sql.NullTime
	var expiresAt time.Time
	err := r.db.conn.QueryRowContext(ctx, `select revoked_at, expires_at from auth_sessions where token=$1`, token).Scan(&revokedAt, &expiresAt)
	if err == sql.ErrNoRows {
		return true, nil
	}
	if err != nil {
		return false, mapError(err)
	}
	return revokedAt.Valid || time.Now().After(expiresAt), nil
}

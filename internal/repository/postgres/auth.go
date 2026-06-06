package postgres

import (
	"context"
	"database/sql"
	"time"
	pkgErr "github.com/Revvfi/revvfi-backend/pkg/errors"
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
@desc Stores an authentication nonce with expiration.
@fix Removed ON CONFLICT clause - allows multiple nonces per wallet since we have composite UNIQUE(wallet_address, nonce)
*/
func (r *AuthRepository) StoreNonce(ctx context.Context, wallet, nonce string, expiresAt time.Time) error {
    _, err := r.db.conn.ExecContext(ctx, `
        INSERT INTO auth_nonces(wallet_address, nonce, expires_at, created_at) 
        VALUES($1, $2, $3, $4)
    `, wallet, nonce, expiresAt, now())
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

/*
@method ValidateNonceExists

@desc
Validates a nonce exists and is NOT consumed (read-only).
Called BEFORE signature verification.

@params
- ctx: request context
- wallet: wallet address
- nonce: nonce to validate

@returns
- error: if nonce doesn't exist, is consumed, or expired
*/
func (r *AuthRepository) ValidateNonceExists(ctx context.Context, wallet, nonce string) error {
	var consumed bool
	var expiresAt time.Time
	
	err := r.db.conn.QueryRowContext(ctx, `
		SELECT consumed, expires_at 
		FROM auth_nonces 
		WHERE wallet_address = $1 AND nonce = $2
	`, wallet, nonce).Scan(&consumed, &expiresAt)
	
	if err == sql.ErrNoRows {
		return pkgErr.ErrNonceNotFound
	}
	if err != nil {
		return mapError(err)
	}
	
	if consumed {
		return pkgErr.ErrNonceAlreadyUsed
	}
	
	if time.Now().After(expiresAt) {
		return pkgErr.ErrNonceExpired
	}
	
	return nil
}

/*
@method ConsumeNonce

@desc
Marks a nonce as consumed (non-atomic version).
Called AFTER successful signature verification.

@params
- ctx: request context
- wallet: wallet address
- nonce: nonce to consume

@returns
- error: if nonce not found or already consumed
*/
func (r *AuthRepository) ConsumeNonce(ctx context.Context, wallet, nonce string) error {
	result, err := r.db.conn.ExecContext(ctx, `
		UPDATE auth_nonces
		SET consumed = TRUE
		WHERE
			wallet_address = $1
			AND nonce = $2
			AND consumed = FALSE
	`, wallet, nonce)
	
	if err != nil {
		return mapError(err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	
	if rows == 0 {
		return pkgErr.ErrNonceNotFound
	}
	
	return nil
}

/*
@method ConsumeNonceAtomic

@desc
Atomically validates and consumes nonce in a single operation.
Prevents race conditions between validation and consumption.

@params
- ctx: request context
- wallet: wallet address
- nonce: nonce to consume

@returns
- error: if nonce not found, already consumed, or expired
*/
func (r *AuthRepository) ConsumeNonceAtomic(ctx context.Context, wallet, nonce string) error {
	result, err := r.db.conn.ExecContext(ctx, `
		UPDATE auth_nonces
		SET consumed = TRUE
		WHERE
			wallet_address = $1
			AND nonce = $2
			AND consumed = FALSE
			AND expires_at > NOW()
	`, wallet, nonce)
	
	if err != nil {
		return mapError(err)
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	
	if rows == 0 {
		var exists bool
		checkErr := r.db.conn.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM auth_nonces WHERE wallet_address = $1 AND nonce = $2)
		`, wallet, nonce).Scan(&exists)
		
		if checkErr == nil && !exists {
			return pkgErr.ErrNonceNotFound
		}

		return pkgErr.ErrNonceAlreadyUsed
	}
	
	return nil
}

/*
@method GetNonce

@desc
Retrieves the most recent nonce for a wallet.

@params
- ctx: request context
- wallet: wallet address

@returns
- string: the nonce
- error: if not found
*/
func (r *AuthRepository) GetNonce(ctx context.Context, wallet string) (string, error) {
	var nonce string
	err := r.db.conn.QueryRowContext(ctx, `
		SELECT nonce 
		FROM auth_nonces 
		WHERE wallet_address = $1 
		ORDER BY created_at DESC 
		LIMIT 1
	`, wallet).Scan(&nonce)
	
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", mapError(err)
	}
	
	return nonce, nil
}

/*
@method DeleteNonce

@desc
Deletes all nonces for a wallet.

@params
- ctx: request context
- wallet: wallet address

@returns
- error: if delete fails
*/
func (r *AuthRepository) DeleteNonce(ctx context.Context, wallet string) error {
	_, err := r.db.conn.ExecContext(ctx, `
		DELETE FROM auth_nonces 
		WHERE wallet_address = $1
	`, wallet)
	return mapError(err)
}

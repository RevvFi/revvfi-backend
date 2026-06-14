package models

import "time"

/*@
file: auth.go
package: models
layer: database

purpose:
Database models used by the authentication and indexer subsystems.

tables:
- auth_nonces
- auth_sessions
- processed_events

notes:
- These are PostgreSQL persistence models.
- These are NOT blockchain event payloads.
- Used by repositories, services, and database queries.
*/

/*@
model: AuthNonce
table: auth_nonces

description:
Stores SIWE (Sign-In With Ethereum) nonces used during wallet
authentication.


*/
type AuthNonce struct {
	ID            int64
	WalletAddress string
	Nonce         string
	Consumed      bool
	ExpiresAt     time.Time
	CreatedAt     time.Time
}

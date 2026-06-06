package auth

import "time"

type SIWEMessage struct {
    Domain         string     `json:"domain"`
    Address        string     `json:"address"`
    Statement      string     `json:"statement,omitempty"`
    URI            string     `json:"uri"`
    Version        string     `json:"version"`
    ChainID        int64      `json:"chainId"`
    Nonce          string     `json:"nonce"`
    IssuedAt       time.Time  `json:"issuedAt"`
    Expiration     *time.Time `json:"expirationTime,omitempty"`
    NotBefore      *time.Time `json:"notBefore,omitempty"`
    RequestID      *string    `json:"requestId,omitempty"`
    Resources      []string   `json:"resources,omitempty"`
    Raw            string     `json:"-"`
    RawLines       []string   `json:"-"`
}

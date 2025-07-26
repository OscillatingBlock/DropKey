package models

import (
	"time"
)

type User struct {
	ID        string `bun:"id,pk" json:"user_id"`
	PublicKey string `bun:"public_key,notnull,unique" json:"public_key"`
}

type Paste struct {
	ID         string    `bun:"id,pk" json:"id"`
	Ciphertext string    `bun:"type:MEDIUMTEXT,notnull" json:"ciphertext"`
	Signature  string    `bun:"signature,notnull" json:"signature"`
	PublicKey  string    `bun:"public_key,notnull" json:"public_key"`
	ExpiresAt  time.Time `bun:"expires_at,notnull" json:"expires_at"`

	User *User `bun:"rel:belongs-to,join:public_key=public_key" json:"-"`
}

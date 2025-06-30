package paste

import (
	"time"

	"Drop-Key/internal/user"
)

type Paste struct {
	ID         string    `bun:"id,pk" json:"id"`
	Ciphertext string    `bun:"ciphertext,notnull" json:"ciphertext"`
	Signature  string    `bun:"signature,notnull" json:"signature"`
	PublicKey  string    `bun:"public_key,notnull" json:"public_key"`
	ExpiresAt  time.Time `bun:"expires_at,notnull" json:"expires_at"`

	User *user.User `bun:"rel:belongs-to,join:public_key=public_key" json:"-"`
}

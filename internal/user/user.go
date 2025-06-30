package user

type User struct {
	ID        string `bun:"id,pk" json:"user_id"`
	PublicKey string `bun:"public_key,notnull,unique" json:"public_key"`
}

package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"greenlight.haodm.net/internal/validator"
)

// Token scope
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl), // Get the current time then + ttl to get the expiry time
		Scope:  scope,
	}

	// Create a slice of random bytes
	randomBytes := make([]byte, 16)

	// Convert randow bytes into text using CSPRNG
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// fast hash
	hash := sha256.Sum256([]byte(token.Plaintext))

	// the Sum256 will return an array not a slice, but the token.Hash is slice type, so hash[:] will convert an array to slice
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (t TokenModel) Insert(token *Token) error {
	query := `
	INSERT INTO tokens(hash, user_id, expiry, scope)
	VALUES($1, $2, $3, $4)`

	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancle := context.WithTimeout(context.Background(), time.Second*3)
	defer cancle()

	_, err := t.DB.ExecContext(ctx, query, args...)

	return err
}

func (t TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = t.Insert(token)

	return token, err
}

func (t TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `DELETE FROM tokens WHERE scope = $1 AND user_id = $2`

	ctx, cancle := context.WithTimeout(context.Background(), time.Second*3)
	defer cancle()

	_, err := t.DB.ExecContext(ctx, query, scope, userID)

	return err
}

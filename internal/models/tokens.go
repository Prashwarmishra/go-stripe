package models

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

const ScopeAuthentication = "authentication"

type Token struct {
	PlainText string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func GenerateToken(userId int, ttl time.Duration, scope string) (*Token, error) {
	token := &Token{
		UserID: int64(userId),
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return nil, err
	}

	token.PlainText = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)
	hash := sha256.Sum256(([]byte(token.PlainText)))
	token.Hash = hash[:]

	return token, nil
}

func (m DBModel) InsertToken(user *User, token *Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Hour)

	defer cancel()

	stmt := `DELETE FROM tokens WHERE user_id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, user.ID)

	if err != nil {
		return err
	}

	stmt = `INSERT INTO tokens (
			name, email, user_id, token_hash, expiry, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err = m.DB.ExecContext(ctx, stmt,
		user.FirstName, user.Email, user.ID, token.Hash, token.Expiry, time.Now(), time.Now(),
	)

	return err
}

func (m *DBModel) GetUserFromToken(token string) (*User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(token))

	stmt := `SELECT u.ID, u.first_name, u.last_name, u.email
					 FROM users u
					 INNER JOIN tokens t ON (u.id = t.user_id)
					 WHERE t.token_hash = ? AND t.expiry > ?`

	user := User{}

	err := m.DB.QueryRowContext(ctx, stmt, tokenHash[:], time.Now()).Scan(
		&user.ID, &user.FirstName, &user.LastName, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

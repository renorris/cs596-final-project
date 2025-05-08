package db

import (
	"context"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	UUID         uuid.UUID
	CreatedAt    time.Time
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
}

type InsertUserContext struct {
	Email             string
	PlaintextPassword string
	FirstName         string
	LastName          string
}

func (p *Pool) InsertUser(ctx context.Context, userCtx *InsertUserContext) (user *User, err error) {
	userUUID, err := uuid.NewRandom()
	if err != nil {
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(userCtx.PlaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return
	}

	user = &User{
		UUID:         userUUID,
		CreatedAt:    time.Now(),
		Email:        userCtx.Email,
		PasswordHash: string(passwordHash),
		FirstName:    userCtx.FirstName,
		LastName:     userCtx.LastName,
	}

	if _, err = p.Exec(ctx, `
		INSERT INTO users
		(uuid, created_at, email, 
		 password, first_name, last_name)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		user.UUID, user.CreatedAt,
		user.Email, user.PasswordHash,
		user.FirstName, user.LastName,
	); err != nil {
		return
	}

	return
}

func (p *Pool) SelectUserByEmail(ctx context.Context, email string) (user *User, err error) {
	row := p.QueryRow(ctx, `
		SELECT 
		uuid, created_at, password, 
		first_name, last_name
		FROM users
		WHERE email = $1;`, email)

	user = &User{}
	if err = row.Scan(
		&user.UUID,
		&user.CreatedAt,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
	); err != nil {
		return
	}
	user.Email = email

	return
}

func (p *Pool) CountNumberOfUsers(ctx context.Context) (count int64, err error) {
	row := p.QueryRow(ctx, `SELECT COUNT(uuid) FROM users;`)
	err = row.Scan(&count)
	return
}

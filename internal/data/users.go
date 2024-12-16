package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/stphngrtz/greenlight/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousUser = &User{}

type User struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

func (u *User) IsAnonymousUser() bool {
	return u == AnonymousUser
}

type password struct {
	plaintext *string // to distinguish between nil and ""
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePassword(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(validator.MinChars(password, 8), "password", "must be at least 8 chars long")
	v.Check(validator.MaxChars(password, 50), "password", "must not be more than 72 chars long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(validator.MaxChars(user.Name, 500), "name", "must not be more than 500 chars long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePassword(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil { // sanity check for own code
		panic("missing password hash for user")
	}
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(user *User) error {
	query := `
		INSERT INTO users (name, email, password_hash, activated)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version
	`

	ctx, close := context.WithTimeout(context.Background(), 3*time.Second)
	defer close()

	err := m.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash, user.Activated).Scan(&user.Id, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*User, error) {
	query := `
		SELECT id, created_at, name, email, password_hash, activated, version
		  FROM users
		 WHERE email = $1
	`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(&user.Id, &user.CreatedAt, &user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) Update(user *User) error {
	query := `
		UPDATE users
		   SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
		 WHERE id = $5 AND version = $6
		 RETURNING version
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, user.Name, user.Email, user.Password.hash, user.Activated, user.Id, user.Version).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenPlaintext, tokenScope string) (*User, error) {
	query := `
		SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
		  FROM users
		 INNER JOIN tokens ON tokens.user_id = users.id
		 WHERE tokens.hash = $1
		   AND tokens.scope = $2
		   AND tokens.expiry > $3
	`

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, tokenHash[:], tokenScope, time.Now()).Scan(&user.Id, &user.CreatedAt, &user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

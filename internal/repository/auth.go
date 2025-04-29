package repository

import (
	"fmt"
	"time"

	"github.com/ShekleinAleksey/auth-service/internal/entity"
	"github.com/jmoiron/sqlx"
)

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}
func (r *AuthRepository) CreateUser(user entity.User) (int, error) {
	var userCreatedID int
	query := fmt.Sprintf("INSERT INTO users (first_name, last_name, email, password_hash) values ($1, $2, $3, $4) RETURNING id")

	row := r.db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.Password)
	if err := row.Scan(&userCreatedID); err != nil {
		return 0, err
	}

	return userCreatedID, nil
}

func (r *AuthRepository) GetUser(email, password string) (entity.User, error) {
	var user entity.User
	query := fmt.Sprintf("SELECT id FROM users WHERE email=$1 AND password_hash=$2")
	err := r.db.Get(&user, query, email, password)

	return user, err
}

func (r *AuthRepository) GetUsers() ([]entity.User, error) {
	var users []entity.User
	query := fmt.Sprintf("SELECT id, first_name, last_name, email FROM users")
	err := r.db.Select(&users, query)

	return users, err
}

func (r *AuthRepository) SaveRefreshToken(userID int, token string, tokenTTL time.Duration) error {
	fmt.Println("userID: ", userID)
	fmt.Println("token: ", token)
	_, err := r.db.Exec("INSERT INTO refresh_tokens (user_id, token_hash, created_at, expires_at) VALUES ($1, $2, $3, $4)", userID, token, time.Now(), time.Now().Add(tokenTTL))
	return err
}

func (r *AuthRepository) FindUserByRefreshToken(userID int) (*entity.User, error) {
	var user entity.User
	err := r.db.QueryRow("SELECT id, first_name, last_name, email FROM users JOIN refresh_tokens ON users.id = refresh_tokens.user_id WHERE refresh_tokens.user_id = $1", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindRefreshToken(userID int) (string, error) {
	var token string
	err := r.db.QueryRow("SELECT token_hash FROM refresh_tokens WHERE user_id = $1", userID).Scan(&token)
	return token, err
}

func (r *AuthRepository) FindUserByID(userID int) (entity.User, error) {
	var user entity.User
	err := r.db.QueryRow("SELECT id, first_name, last_name, email FROM users WHERE id = $1", userID).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email)
	return user, err
}

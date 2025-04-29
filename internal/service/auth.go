package service

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"time"

	"github.com/ShekleinAleksey/auth-service/internal/entity"
	"github.com/ShekleinAleksey/auth-service/internal/repository"
	"github.com/dgrijalva/jwt-go"
)

type AuthService struct {
	repo repository.AuthRepository
}

func NewAuthService(repo repository.AuthRepository) *AuthService {
	return &AuthService{repo: repo}
}

const (
	salt       = "hjqrhjqw124617ajfhajs"
	signingKey = "qrkjk#4#%35FSFJlja#4353KSFjH"
	tokenTTL   = 12 * time.Hour
	refreshTTL = 30 * 24 * time.Hour
)

type tokenClaims struct {
	jwt.StandardClaims
	UserId int `json:"user_id"`
}

func (s *AuthService) CreateUser(user entity.User) (int, error) {
	user.Password = generatePasswordHash(user.Password)
	userID, err := s.repo.CreateUser(user)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *AuthService) CreateToken(email, password string) (TokenDetails, error) {
	user, err := s.repo.GetUser(email, generatePasswordHash(password))
	if err != nil {
		return TokenDetails{}, errors.New("user not found")
	}

	tokenDetails, err := s.GenerateToken(user)
	if err != nil {
		return TokenDetails{}, err
	}

	return tokenDetails, nil
}

func (s *AuthService) GenerateToken(user entity.User) (TokenDetails, error) {

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})
	fmt.Println("accessToken: ", accessToken)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(refreshTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		user.ID,
	})
	accessTokenString, err := accessToken.SignedString([]byte(signingKey))
	if err != nil {
		return TokenDetails{}, err
	}

	refreshTokenString, err := refreshToken.SignedString([]byte(signingKey))
	if err != nil {
		return TokenDetails{}, err
	}

	err = s.repo.SaveRefreshToken(user.ID, refreshTokenString, refreshTTL)
	if err != nil {
		return TokenDetails{}, err
	}

	claims, err := s.ParseToken(accessTokenString)
	if err != nil {
		return TokenDetails{}, err
	}
	expiresIn := claims.ExpiresAt
	claimsRefresh, err := s.ParseToken(refreshTokenString)
	if err != nil {
		return TokenDetails{}, err
	}
	refreshExpiresIn := claimsRefresh.ExpiresAt

	return TokenDetails{
		AccessToken:      accessTokenString,
		RefreshToken:     refreshTokenString,
		ExpiresIn:        expiresIn,
		RefreshExpiresIn: refreshExpiresIn,
	}, nil
}

type TokenDetails struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}

func (s *AuthService) ParseToken(accessToken string) (*tokenClaims, error) {
	token, err := jwt.ParseWithClaims(accessToken, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}

		return []byte(signingKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok {
		return nil, errors.New("token claims are not of type *tokenClaims")
	}

	return claims, nil
}

func generatePasswordHash(password string) string {
	hash := sha1.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(salt)))
}

func (s *AuthService) GetUsers() ([]entity.User, error) {
	return s.repo.GetUsers()
}

func (s *AuthService) FindRefreshToken(userID int) (string, error) {
	refreshToken, err := s.repo.FindRefreshToken(userID)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *AuthService) FindUser(userID int) (entity.User, error) {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return entity.User{}, err
	}

	return user, nil
}

func (s *AuthService) RefreshToken(refreshToken string) (TokenDetails, error) {
	claims, err := s.ParseToken(refreshToken)
	userID := claims.UserId
	if err != nil {
		return TokenDetails{}, err
	}

	user, err := s.FindUser(userID)
	if err != nil {
		return TokenDetails{}, err
	}

	savedRefreshToken, err := s.FindRefreshToken(userID)
	if err != nil {
		return TokenDetails{}, err
	}

	if savedRefreshToken != refreshToken {
		return TokenDetails{}, errors.New("savedRefreshToken != refreshToken")
	}

	tokenDetails, err := s.GenerateToken(user)
	if err != nil {
		return TokenDetails{}, err
	}

	return tokenDetails, nil
}

// func createToken(guid, ip string) (*TokenDetails, error) {
// 	accessTokenExpiration := time.Now().Add(30 * time.Minute).Unix()
// 	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"user_id": guid,
// 		"ip":      ip,
// 		"exp":     accessTokenExpiration,
// 	})

// 	accessTokenString, err := accessToken.SignedString([]byte("your_secret_key"))
// 	if err != nil {
// 		return nil, err
// 	}

// 	refreshToken := uuid.New().String()

// 	if err := saveRefreshToken(guid, refreshToken); err != nil {
// 		return nil, err
// 	}

// 	return &TokenDetails{
// 		AccessToken:  accessTokenString,
// 		RefreshToken: refreshToken,
// 		ExpiresAt:    accessTokenExpiration,
// 	}, nil
// }

package service

import "github.com/ShekleinAleksey/auth-service/internal/repository"

type Service struct {
	AuthService *AuthService
	//sUserService *UserService
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		AuthService: NewAuthService(*repo.AuthRepository),
		//UserService: NewUserService(*repo.UserRepository),
	}
}

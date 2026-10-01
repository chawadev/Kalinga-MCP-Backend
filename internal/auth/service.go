package auth

import (
	"context"
	"errors"

	"github.com/chawadev/kalinga-backend/internal/users"
)

type Service struct {
	userRepo *users.Repository
}

func NewService(userRepo *users.Repository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) Register(ctx context.Context, req *users.RegisterRequest) (*users.AuthResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &users.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// Generate token
	token, err := GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &users.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) Login(ctx context.Context, req *users.LoginRequest) (*users.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}

	token, err := GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &users.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*users.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

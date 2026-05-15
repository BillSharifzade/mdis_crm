package service

import (
	"context"
	"fmt"
	"unicode"

	"crm_backend/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// validatePassword — минимальная политика: длина ≥ 8, есть буква и цифра.
func validatePassword(p string) error {
	if len(p) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	var hasLetter, hasDigit bool
	for _, r := range p {
		if unicode.IsLetter(r) {
			hasLetter = true
		}
		if unicode.IsDigit(r) {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return fmt.Errorf("password must contain both letters and digits")
	}
	return nil
}

func (s *UserService) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.repo.ListAll(ctx)
}

func (s *UserService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.User, error) {
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	return s.repo.CreateUser(ctx, req.Name, req.Email, string(hash), req.Role)
}

func (s *UserService) UpdateUser(ctx context.Context, id int, req *model.UpdateUserRequest) (*model.User, error) {
	passwordHash := ""
	if req.Password != "" {
		if err := validatePassword(req.Password); err != nil {
			return nil, err
		}
		h, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		passwordHash = string(h)
	}
	return s.repo.UpdateUser(ctx, id, req.Name, req.Email, req.Role, passwordHash)
}

func (s *UserService) DeleteUser(ctx context.Context, id int) error {
	return s.repo.DeleteUser(ctx, id)
}

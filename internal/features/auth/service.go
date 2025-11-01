package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/maevlava/resume-backend/internal/shared/common"
	"github.com/maevlava/resume-backend/internal/shared/config"
	"github.com/maevlava/resume-backend/internal/shared/db"
	"github.com/maevlava/resume-backend/internal/shared/domain"
)

var (
	ErrEmailExists = errors.New("email already exists")
)

type Service struct {
	cfg *config.Config
	db  *db.Queries
}

func NewService(cfg *config.Config, db *db.Queries) *Service {
	return &Service{
		cfg: cfg,
		db:  db,
	}
}
func (s *Service) Login(ctx context.Context, email, password string) (*common.Token, error) {
	// exist
	user, err := s.db.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("login: user not found")
		}
		return nil, fmt.Errorf("login: failed to get user by email: %w", err)
	}

	// password match
	err = common.CheckPasswordHash(password, user.HashedPassword)
	if err != nil {
		return nil, fmt.Errorf("login: invalid credentials: %w", err)
	}

	// generate JWT token
	token, err := common.MakeJWT(user, s.cfg.JWTSecret, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("login: failed to generate token: %w", err)
	}

	return token, nil
}
func (s *Service) Register(ctx context.Context, username, email, password string) error {

	hashedPassword, err := common.HashPassword(password)
	if err != nil {
		return fmt.Errorf("register: failed to hash password: %w", err)
	}
	dbParams := db.CreateUserParams{
		ID:             uuid.New(),
		Name:           username,
		Email:          email,
		HashedPassword: hashedPassword,
	}

	_, err = s.db.CreateUser(ctx, dbParams)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrEmailExists
		}
		return fmt.Errorf("register: failed to create user: %w", err)
	}

	return nil

}
func (s *Service) Validate(token string) error {
	_, err := common.ValidateJWT(token, s.cfg.JWTSecret)
	if err != nil {
		return fmt.Errorf("validate: failed to validate token: %w", err)
	}
	return nil
}
func (s *Service) GetSignedInUser(ctx context.Context, token string) (*domain.User, error) {
	claims, err := common.ValidateJWT(token, s.cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("getSignedInUser: failed to validate token: %w", err)
	}

	userIDString, err := claims.GetSubject()
	if err != nil {
		return nil, fmt.Errorf("getSignedInUser: failed to get subject: %w", err)
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return nil, fmt.Errorf("getSignedInUser: failed to parse subject: %w", err)
	}

	user, err := s.db.GetUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getSignedInUser: failed to get user: %w", err)
	}

	domainUser := domain.User{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return &domainUser, nil
}

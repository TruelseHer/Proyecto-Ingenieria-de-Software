package service

import (
	"fmt"
	"strings"
	"time"
	"trotamundos/internal/domain"
	"trotamundos/internal/platform"
	"trotamundos/internal/repository"
)

type AuthService struct {
	store  *repository.Store
	tokens *platform.TokenService
}

func NewAuthService(store *repository.Store, tokens *platform.TokenService) *AuthService {
	return &AuthService{
		store:  store,
		tokens: tokens,
	}
}

func (s *AuthService) Register(name, email, password, homeCurrency, preferredAgencyID string) (domain.User, string, error) {
	name = strings.TrimSpace(name)
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	homeCurrency = strings.ToUpper(strings.TrimSpace(homeCurrency))
	preferredAgencyID = strings.TrimSpace(preferredAgencyID)

	if name == "" || email == "" || !strings.Contains(email, "@") {
		return domain.User{}, "", fmt.Errorf("%w: ingresa tu nombre completo y un correo electrónico válido", domain.ErrValidation)
	}
	if len(password) < 6 {
		return domain.User{}, "", fmt.Errorf("%w: la contraseña debe tener al menos 6 caracteres", domain.ErrValidation)
	}

	if _, err := s.store.GetUserByEmail(email); err == nil {
		return domain.User{}, "", domain.ErrEmailAlreadyUsed
	}

	hash, err := platform.HashPassword(password)
	if err != nil {
		return domain.User{}, "", err
	}

	if homeCurrency == "" {
		homeCurrency = "USD"
	}
	if preferredAgencyID == "" {
		agencies := s.store.ListAgencies()
		if len(agencies) > 0 {
			preferredAgencyID = agencies[0].ID
		} else {
			preferredAgencyID = "ag-meridian"
		}
	}

	u := domain.User{
		ID:                platform.NewID("usr"),
		Name:              name,
		Email:             email,
		PasswordHash:      hash,
		Role:              "traveler",
		HomeCurrency:      homeCurrency,
		PreferredAgencyID: preferredAgencyID,
		MailboxAuthorized: true,
		ConnectedEmail:    email,
		AvatarInitials:    computeInitials(name),
		CreatedAt:         time.Now().UTC(),
	}

	if err := s.store.SaveUser(u); err != nil {
		return domain.User{}, "", err
	}
	s.store.EnsureDefaultSenderRules(u.ID)

	tok, err := s.tokens.Issue(u.ID, u.Role, 24*time.Hour)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, tok, nil
}

func (s *AuthService) Login(email, password string) (domain.User, string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	password = strings.TrimSpace(password)
	if email == "" || password == "" {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}

	u, err := s.store.GetUserByEmail(email)
	if err != nil {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}
	if !platform.VerifyPassword(password, u.PasswordHash) {
		return domain.User{}, "", domain.ErrInvalidCredentials
	}

	s.store.EnsureDefaultSenderRules(u.ID)

	tok, err := s.tokens.Issue(u.ID, u.Role, 24*time.Hour)
	if err != nil {
		return domain.User{}, "", err
	}
	return u, tok, nil
}

func (s *AuthService) Logout(token string) {
	if s.tokens != nil && token != "" {
		s.tokens.Revoke(token, 24*time.Hour)
	}
}

func computeInitials(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "TM"
	}
	if len(parts) == 1 {
		r := []rune(parts[0])
		if len(r) >= 2 {
			return strings.ToUpper(string(r[:2]))
		}
		return strings.ToUpper(string(r))
	}
	first := []rune(parts[0])
	last := []rune(parts[len(parts)-1])
	return strings.ToUpper(string(first[0]) + string(last[0]))
}

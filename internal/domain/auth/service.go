package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo      user.Repository
	blacklistRepo TokenBlacklistRepository
	tokenService  TokenService
	hasher        PasswordHasher
}

func NewAuthService(
	userRepo user.Repository,
	blacklistRepo TokenBlacklistRepository,
	tokenService TokenService,
	hasher PasswordHasher,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		blacklistRepo: blacklistRepo,
		tokenService:  tokenService,
		hasher:        hasher,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	existing, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, user.ErrEmailAlreadyExists
	}

	hashedPassword, err := s.hasher.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser, err := user.NewUser(req.FirstName, req.LastName, email, hashedPassword)
	if err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to persist user: %w", err)
	}

	tokenPair, _, _, err := s.tokenService.GenerateTokenPair(newUser.ID, newUser.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         newUser,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	userEntity, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if !userEntity.IsActive() {
		return nil, user.ErrUserInactive
	}

	if err := s.hasher.Compare(userEntity.Password, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	tokenPair, _, _, err := s.tokenService.GenerateTokenPair(userEntity.ID, userEntity.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         userEntity,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req dto.RefreshRequest) (*dto.AuthResponse, error) {
	claims, err := s.tokenService.ValidateToken(req.RefreshToken, TokenTypeRefresh)
	if err != nil {
		return nil, err
	}

	isBlacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, ErrTokenRevoked
	}

	// Rotate refresh token
	_ = s.blacklistRepo.BlacklistToken(ctx, claims.JTI, claims.UserID, TokenTypeRefresh, claims.ExpiresAt, "refreshed")

	userEntity, err := s.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, user.ErrUserNotFound
	}

	if !userEntity.IsActive() {
		return nil, user.ErrUserInactive
	}

	tokenPair, _, _, err := s.tokenService.GenerateTokenPair(userEntity.ID, userEntity.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	return &dto.AuthResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		TokenType:    tokenPair.TokenType,
		ExpiresIn:    tokenPair.ExpiresIn,
		User:         userEntity,
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, req dto.LogoutRequest) error {
	if req.AccessToken != "" {
		claims, err := s.tokenService.ValidateToken(req.AccessToken, TokenTypeAccess)
		if err == nil && claims != nil {
			_ = s.blacklistRepo.BlacklistToken(ctx, claims.JTI, claims.UserID, TokenTypeAccess, claims.ExpiresAt, "logout")
		}
	}

	if req.RefreshToken != "" {
		claims, err := s.tokenService.ValidateToken(req.RefreshToken, TokenTypeRefresh)
		if err == nil && claims != nil {
			_ = s.blacklistRepo.BlacklistToken(ctx, claims.JTI, claims.UserID, TokenTypeRefresh, claims.ExpiresAt, "logout")
		}
	}

	return nil
}

func (s *AuthService) ValidateAccessToken(ctx context.Context, tokenStr string) (*TokenClaims, error) {
	claims, err := s.tokenService.ValidateToken(tokenStr, TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	isBlacklisted, err := s.blacklistRepo.IsTokenBlacklisted(ctx, claims.JTI)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, ErrTokenRevoked
	}

	return claims, nil
}

type APIKeyService struct {
	apiKeyRepo   APIKeyRepository
	keyGenerator *APIKeyGenerator
}

func NewAPIKeyService(apiKeyRepo APIKeyRepository, keyGenerator *APIKeyGenerator) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo:   apiKeyRepo,
		keyGenerator: keyGenerator,
	}
}

func (s *APIKeyService) CreateAPIKey(ctx context.Context, orgID, projectID, userID uuid.UUID, req dto.CreateAPIKeyRequest) (*dto.CreatedAPIKeyResponse, error) {
	generated, err := s.keyGenerator.Generate()
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure key: %w", err)
	}

	apiKeyEntity := &APIKey{
		ID:             uuid.New(),
		OrganizationID: orgID,
		ProjectID:      projectID,
		UserID:         userID,
		Name:           req.Name,
		KeyPrefix:      generated.KeyPrefix,
		KeyHash:        generated.KeyHash,
		Status:         APIKeyStatusActive,
		ExpiresAt:      req.ExpiresAt,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	if err := s.apiKeyRepo.Create(ctx, apiKeyEntity); err != nil {
		return nil, fmt.Errorf("failed to store API key: %w", err)
	}

	return &dto.CreatedAPIKeyResponse{
		ID:        apiKeyEntity.ID,
		Name:      apiKeyEntity.Name,
		SecretKey: generated.PlaintextKey,
		KeyPrefix: generated.KeyPrefix,
		ExpiresAt: apiKeyEntity.ExpiresAt,
		CreatedAt: apiKeyEntity.CreatedAt,
	}, nil
}

func (s *APIKeyService) RevokeAPIKey(ctx context.Context, keyID uuid.UUID) error {
	return s.apiKeyRepo.Revoke(ctx, keyID)
}

func (s *APIKeyService) ListProjectAPIKeys(ctx context.Context, projectID uuid.UUID) ([]*dto.APIKeyResponse, error) {
	keys, err := s.apiKeyRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.APIKeyResponse, len(keys))
	for i, k := range keys {
		responses[i] = &dto.APIKeyResponse{
			ID:             k.ID,
			OrganizationID: k.OrganizationID,
			ProjectID:      k.ProjectID,
			UserID:         k.UserID,
			Name:           k.Name,
			KeyPrefix:      k.KeyPrefix,
			Status:         string(k.Status),
			ExpiresAt:      k.ExpiresAt,
			LastUsedAt:     k.LastUsedAt,
			CreatedAt:      k.CreatedAt,
		}
	}
	return responses, nil
}

func (s *APIKeyService) AuthenticateAPIKey(ctx context.Context, plaintextKey string) (*dto.AuthenticatedPrincipal, error) {
	if plaintextKey == "" {
		return nil, ErrAPIKeyNotFound
	}

	hash := HashAPIKey(plaintextKey)
	key, err := s.apiKeyRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, ErrAPIKeyNotFound
	}

	if !key.IsActive() {
		return nil, ErrAPIKeyInactive
	}

	_ = s.apiKeyRepo.UpdateLastUsed(ctx, key.ID)

	return &dto.AuthenticatedPrincipal{
		APIKeyID:       key.ID,
		OrganizationID: key.OrganizationID,
		ProjectID:      key.ProjectID,
		UserID:         key.UserID,
		KeyPrefix:      key.KeyPrefix,
	}, nil
}

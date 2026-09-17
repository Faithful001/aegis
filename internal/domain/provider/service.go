package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/provider/dto"
	"github.com/Faithful001/aegis/internal/pkg/crypto"
	"github.com/google/uuid"
)

type ProviderCredentialService struct {
	repo          IProviderCredentialRepository
	encryptionKey []byte
}

func NewProviderCredentialService(repo IProviderCredentialRepository, encryptionKey string) *ProviderCredentialService {
	keyBytes := []byte(encryptionKey)
	if len(keyBytes) == 0 {
		keyBytes = []byte("aegis-default-byok-secret-key-32b")
	}
	return &ProviderCredentialService{
		repo:          repo,
		encryptionKey: keyBytes,
	}
}

func (s *ProviderCredentialService) SaveCredential(
	ctx context.Context,
	orgID uuid.UUID,
	req dto.SaveCredentialRequest,
) (*dto.CredentialResponse, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if !IsSupportedProvider(provider) {
		return nil, ErrInvalidProvider
	}

	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		return nil, ErrInvalidCredentialData
	}

	encryptedKey, err := crypto.AESEncrypt(apiKey, s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %w", err)
	}

	cred := &ProviderCredential{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		Provider:        provider,
		EncryptedAPIKey: encryptedKey,
		BaseURL:         strings.TrimSpace(req.BaseURL),
	}

	if err := s.repo.Save(ctx, cred); err != nil {
		return nil, err
	}

	return &dto.CredentialResponse{
		ID:             cred.ID,
		OrganizationID: cred.OrganizationID,
		Provider:       cred.Provider,
		MaskedAPIKey:   crypto.MaskAPIKey(apiKey),
		BaseURL:        cred.BaseURL,
		CreatedAt:      cred.CreatedAt,
		UpdatedAt:      cred.UpdatedAt,
	}, nil
}

func (s *ProviderCredentialService) GetDecryptedKey(
	ctx context.Context,
	orgID uuid.UUID,
	provider string,
) (apiKey string, baseURL string, err error) {
	cred, err := s.repo.GetByOrgAndProvider(ctx, orgID, provider)
	if err != nil {
		return "", "", err
	}

	decryptedKey, err := crypto.AESDecrypt(cred.EncryptedAPIKey, s.encryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return decryptedKey, cred.BaseURL, nil
}

func (s *ProviderCredentialService) ListCredentials(
	ctx context.Context,
	orgID uuid.UUID,
) (*dto.CredentialListResponse, error) {
	creds, err := s.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	resp := make([]dto.CredentialResponse, len(creds))
	for i, c := range creds {
		maskedKey := "****"
		if decryptedKey, err := crypto.AESDecrypt(c.EncryptedAPIKey, s.encryptionKey); err == nil {
			maskedKey = crypto.MaskAPIKey(decryptedKey)
		}

		resp[i] = dto.CredentialResponse{
			ID:             c.ID,
			OrganizationID: c.OrganizationID,
			Provider:       c.Provider,
			MaskedAPIKey:   maskedKey,
			BaseURL:        c.BaseURL,
			CreatedAt:      c.CreatedAt,
			UpdatedAt:      c.UpdatedAt,
		}
	}

	return &dto.CredentialListResponse{Data: resp}, nil
}

func (s *ProviderCredentialService) DeleteCredential(
	ctx context.Context,
	orgID uuid.UUID,
	provider string,
) error {
	if !IsSupportedProvider(provider) {
		return ErrInvalidProvider
	}
	return s.repo.Delete(ctx, orgID, provider)
}

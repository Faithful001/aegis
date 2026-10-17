package provider

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Faithful001/aegis/internal/domain/provider/dto"
	"github.com/Faithful001/aegis/internal/pkg/crypto"
	"github.com/google/uuid"
)

type ProviderModelLister interface {
	ListModels(ctx context.Context, providerName, apiKey, baseURL string) ([]dto.ModelDescriptor, error)
}

type ProviderCredentialService struct {
	repo          IProviderCredentialRepository
	encryptionKey []byte
	gateway       ProviderModelLister
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

func (s *ProviderCredentialService) SetGateway(gw ProviderModelLister) {
	s.gateway = gw
}

func (s *ProviderCredentialService) SaveCredential(
	ctx context.Context,
	userID uuid.UUID,
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
		UserID:          userID,
		Provider:        provider,
		EncryptedAPIKey: encryptedKey,
		BaseURL:         strings.TrimSpace(req.BaseURL),
	}

	if err := s.repo.Save(ctx, cred); err != nil {
		return nil, err
	}

	return &dto.CredentialResponse{
		ID:           cred.ID,
		UserID:       cred.UserID,
		Provider:     cred.Provider,
		MaskedAPIKey: crypto.MaskAPIKey(apiKey),
		BaseURL:      cred.BaseURL,
		CreatedAt:    cred.CreatedAt,
		UpdatedAt:    cred.UpdatedAt,
	}, nil
}

func (s *ProviderCredentialService) GetDecryptedKey(
	ctx context.Context,
	userID uuid.UUID,
	provider string,
) (apiKey string, baseURL string, err error) {
	cred, err := s.repo.GetByUserAndProvider(ctx, userID, provider)
	if err != nil {
		return "", "", err
	}

	log.Printf("User ID: %v", userID)
	log.Printf("Provider: %v", provider)

	decryptedKey, err := crypto.AESDecrypt(cred.EncryptedAPIKey, s.encryptionKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt API key: %w", err)
	}

	return decryptedKey, cred.BaseURL, nil
}

func (s *ProviderCredentialService) ListCredentials(
	ctx context.Context,
	userID uuid.UUID,
) (*dto.CredentialListResponse, error) {
	creds, err := s.repo.ListByUser(ctx, userID)
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
			ID:           c.ID,
			UserID:       c.UserID,
			Provider:     c.Provider,
			MaskedAPIKey: maskedKey,
			BaseURL:      c.BaseURL,
			CreatedAt:    c.CreatedAt,
			UpdatedAt:    c.UpdatedAt,
		}
	}

	return &dto.CredentialListResponse{Data: resp}, nil
}

func (s *ProviderCredentialService) DeleteCredential(
	ctx context.Context,
	userID uuid.UUID,
	provider string,
) error {
	if !IsSupportedProvider(provider) {
		return ErrInvalidProvider
	}
	return s.repo.Delete(ctx, userID, provider)
}

func (s *ProviderCredentialService) ListProviderModels(
	ctx context.Context,
	userID uuid.UUID,
	providerName string,
) ([]dto.ModelDescriptor, error) {
	if s.gateway == nil {
		return nil, errors.New("provider gateway not initialized")
	}

	p := strings.ToLower(strings.TrimSpace(providerName))
	if !IsSupportedProvider(p) {
		return nil, ErrInvalidProvider
	}

	apiKey, baseURL, err := s.GetDecryptedKey(ctx, userID, p)
	if err != nil {
		return nil, err
	}

	gatewayModels, err := s.gateway.ListModels(ctx, p, apiKey, baseURL)
	if err != nil {
		return nil, err
	}

	// Fetch existing user models for this provider from DB
	dbModels, _ := s.repo.ListUserModels(ctx, userID, p, false)
	activeMap := make(map[string]bool)
	for _, m := range dbModels {
		activeMap[m.ModelID] = m.IsActive
	}

	toCreate := make([]*UserModel, 0)
	result := make([]dto.ModelDescriptor, len(gatewayModels))
	for i, m := range gatewayModels {
		isActive := false
		if val, exists := activeMap[m.ID]; exists {
			isActive = val
		} else {
			// Newly discovered model - persist to DB with default inactive
			toCreate = append(toCreate, &UserModel{
				ID:          uuid.New(),
				UserID:      userID,
				Provider:    p,
				ModelID:     m.ID,
				Name:        m.Name,
				Description: m.Description,
				Fast:        m.Fast,
				IsActive:    false,
			})
		}

		result[i] = dto.ModelDescriptor{
			ID:          m.ID,
			Name:        m.Name,
			Provider:    p,
			Description: m.Description,
			Fast:        m.Fast,
			IsActive:    isActive,
		}
	}

	if len(toCreate) > 0 {
		_ = s.repo.UpsertUserModelsBatch(ctx, toCreate)
	}

	return result, nil
}

func (s *ProviderCredentialService) ListAllModels(
	ctx context.Context,
	userID uuid.UUID,
	isActiveFilter ...*bool,
) ([]dto.ModelDescriptor, error) {
	var filterActive *bool
	if len(isActiveFilter) > 0 {
		filterActive = isActiveFilter[0]
	}

	// When active models are requested, query directly from the DB level!
	if filterActive != nil && *filterActive {
		dbModels, err := s.repo.ListAllUserModels(ctx, userID, true)
		if err != nil {
			return nil, err
		}

		result := make([]dto.ModelDescriptor, len(dbModels))
		for i, m := range dbModels {
			result[i] = dto.ModelDescriptor{
				ID:          m.ModelID,
				Name:        m.Name,
				Provider:    m.Provider,
				Description: m.Description,
				Fast:        m.Fast,
				IsActive:    m.IsActive,
			}
		}
		return result, nil
	}

	allModels := make([]dto.ModelDescriptor, 0)
	creds, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	for _, c := range creds {
		models, err := s.ListProviderModels(ctx, userID, c.Provider)
		if err != nil {
			continue
		}
		for _, m := range models {
			if !seen[m.ID] {
				seen[m.ID] = true
				if filterActive != nil && m.IsActive != *filterActive {
					continue
				}
				allModels = append(allModels, m)
			}
		}
	}

	return allModels, nil
}

func (s *ProviderCredentialService) ToggleModelActive(
	ctx context.Context,
	userID uuid.UUID,
	providerName string,
	modelID string,
	isActive *bool,
) (*dto.ModelDescriptor, error) {
	p := strings.ToLower(strings.TrimSpace(providerName))
	if !IsSupportedProvider(p) {
		return nil, ErrInvalidProvider
	}

	updated, err := s.repo.ToggleUserModel(ctx, userID, p, modelID, isActive)
	if err != nil {
		return nil, err
	}

	return &dto.ModelDescriptor{
		ID:          updated.ModelID,
		Name:        updated.Name,
		Provider:    updated.Provider,
		Description: updated.Description,
		Fast:        updated.Fast,
		IsActive:    updated.IsActive,
	}, nil
}

// staticProviders returns the immutable provider catalogue.
func staticProviders() []dto.ProviderDescriptor {
	return []dto.ProviderDescriptor{
		{
			ID:          "openai",
			Name:        "OpenAI",
			Description: "GPT models via api.openai.com",
			Placeholder: "sk-proj-...",
			KeyLink:     "https://platform.openai.com/api-keys",
		},
		{
			ID:          "anthropic",
			Name:        "Anthropic",
			Description: "Claude models via api.anthropic.com",
			Placeholder: "sk-ant-...",
			KeyLink:     "https://console.anthropic.com/account/keys",
		},
		{
			ID:          "gemini",
			Name:        "Google Gemini",
			Description: "Gemini models via Google AI Studio",
			Placeholder: "AIzaSy...",
			KeyLink:     "https://aistudio.google.com/app/apikey",
		},
		{
			ID:          "mistral",
			Name:        "Mistral",
			Description: "Mistral models via api.mistral.ai",
			Placeholder: "your_mistral_api_key",
			KeyLink:     "https://console.mistral.ai/api-keys/",
		},
		{
			ID:          "openrouter",
			Name:        "OpenRouter",
			Description: "Universal router for 200+ models",
			Placeholder: "sk-or-...",
			KeyLink:     "https://openrouter.ai/keys",
		},
	}
}

// GetSupportedProviders returns the provider catalogue annotated with the
// user's credential status (IsConfigured + MaskedAPIKey).
func (s *ProviderCredentialService) GetSupportedProviders(ctx context.Context, userID uuid.UUID) []dto.ProviderDescriptor {
	providers := staticProviders()

	// Build a map of provider → masked key from the user's stored credentials.
	credMap := make(map[string]string)
	if userID != uuid.Nil {
		if creds, err := s.repo.ListByUser(ctx, userID); err == nil {
			for _, c := range creds {
				masked := "••••••••"
				if decrypted, err := crypto.AESDecrypt(c.EncryptedAPIKey, s.encryptionKey); err == nil {
					masked = crypto.MaskAPIKey(decrypted)
				}
				credMap[c.Provider] = masked
			}
		}
	}

	for i, p := range providers {
		if masked, ok := credMap[p.ID]; ok {
			providers[i].IsConfigured = true
			providers[i].MaskedAPIKey = &masked
		} else {
			providers[i].IsConfigured = false
			providers[i].MaskedAPIKey = nil
		}
	}
	return providers
}

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

	return s.gateway.ListModels(ctx, p, apiKey, baseURL)
}

func (s *ProviderCredentialService) ListAllModels(
	ctx context.Context,
	userID uuid.UUID,
) ([]dto.ModelDescriptor, error) {
	allModels := make([]dto.ModelDescriptor, 0)

	creds, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)

	if s.gateway != nil && len(creds) > 0 {
		for _, c := range creds {
			decryptedKey, err := crypto.AESDecrypt(c.EncryptedAPIKey, s.encryptionKey)
			if err != nil {
				continue
			}
			models, err := s.gateway.ListModels(ctx, c.Provider, decryptedKey, c.BaseURL)
			if err != nil {
				continue
			}
			for _, m := range models {
				if !seen[m.ID] {
					seen[m.ID] = true
					allModels = append(allModels, m)
				}
			}
		}
	}

	return allModels, nil
}

func (s *ProviderCredentialService) GetSupportedProviders() []dto.ProviderDescriptor {
	return []dto.ProviderDescriptor{
		{
			ID:          "openai",
			Name:        "OpenAI",
			Description: "GPT models via api.openai.com",
			Placeholder: "sk-proj-...",
			KeyLink:     "https://platform.openai.com/api-keys",
			Icon:        "◎",
			IconStyle:   "bg-emerald-500/10 text-emerald-400",
		},
		{
			ID:          "anthropic",
			Name:        "Anthropic",
			Description: "Claude models via api.anthropic.com",
			Placeholder: "sk-ant-...",
			KeyLink:     "https://console.anthropic.com/account/keys",
			Icon:        "A",
			IconStyle:   "bg-[#CC785C]/15 text-[#CC785C] font-bold",
		},
		{
			ID:          "gemini",
			Name:        "Google Gemini",
			Description: "Gemini models via Google AI Studio",
			Placeholder: "AIzaSy...",
			KeyLink:     "https://aistudio.google.com/app/apikey",
			Icon:        "✦",
			IconStyle:   "bg-blue-500/10 text-blue-400",
		},
		{
			ID:          "mistral",
			Name:        "Mistral",
			Description: "Mistral models via api.mistral.ai",
			Placeholder: "your_mistral_api_key",
			KeyLink:     "https://console.mistral.ai/api-keys/",
			Icon:        "M",
			IconStyle:   "bg-orange-500/10 text-orange-400 font-bold",
		},
		{
			ID:          "openrouter",
			Name:        "OpenRouter",
			Description: "Universal router for 200+ models",
			Placeholder: "sk-or-...",
			KeyLink:     "https://openrouter.ai/keys",
			Icon:        "⊙",
			IconStyle:   "bg-purple-500/10 text-purple-400",
		},
	}
}

package auth_test

import (
	"context"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/auth"
	authDto "github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/google/uuid"
)

type mockAPIKeyRepo struct {
	keys   map[uuid.UUID]*auth.APIKey
	byHash map[string]*auth.APIKey
}

func newMockAPIKeyRepo() *mockAPIKeyRepo {
	return &mockAPIKeyRepo{
		keys:   make(map[uuid.UUID]*auth.APIKey),
		byHash: make(map[string]*auth.APIKey),
	}
}

func (m *mockAPIKeyRepo) Create(ctx context.Context, key *auth.APIKey) error {
	m.keys[key.ID] = key
	m.byHash[key.KeyHash] = key
	return nil
}

func (m *mockAPIKeyRepo) GetByID(ctx context.Context, id uuid.UUID) (*auth.APIKey, error) {
	k, exists := m.keys[id]
	if !exists {
		return nil, auth.ErrAPIKeyNotFound
	}
	return k, nil
}

func (m *mockAPIKeyRepo) GetByHash(ctx context.Context, hash string) (*auth.APIKey, error) {
	k, exists := m.byHash[hash]
	if !exists {
		return nil, auth.ErrAPIKeyNotFound
	}
	return k, nil
}

func (m *mockAPIKeyRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*auth.APIKey, error) {
	var list []*auth.APIKey
	for _, k := range m.keys {
		if k.ProjectID == projectID {
			list = append(list, k)
		}
	}
	return list, nil
}

func (m *mockAPIKeyRepo) Update(ctx context.Context, key *auth.APIKey) error {
	m.keys[key.ID] = key
	m.byHash[key.KeyHash] = key
	return nil
}

func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *mockAPIKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	if k, exists := m.keys[id]; exists {
		k.Status = auth.APIKeyStatusRevoked
	}
	return nil
}

func (m *mockAPIKeyRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if k, exists := m.keys[id]; exists {
		delete(m.byHash, k.KeyHash)
		delete(m.keys, id)
	}
	return nil
}

func TestAPIKeyService_Lifecycle(t *testing.T) {
	apiKeyRepo := newMockAPIKeyRepo()
	generator := auth.NewAPIKeyGenerator()

	svc := auth.NewAPIKeyService(apiKeyRepo, generator)
	ctx := context.Background()

	orgID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()

	// 1. Create API Key
	req := authDto.CreateAPIKeyRequest{
		Name: "Production Secret Key",
	}
	created, err := svc.CreateAPIKey(ctx, orgID, projectID, userID, req)
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	if created.SecretKey == "" {
		t.Fatalf("expected secret key in creation response")
	}

	// 2. Authenticate with Plaintext Key
	principal, err := svc.AuthenticateAPIKey(ctx, created.SecretKey)
	if err != nil {
		t.Fatalf("AuthenticateAPIKey failed: %v", err)
	}
	if principal.ProjectID != projectID || principal.OrganizationID != orgID {
		t.Fatalf("incorrect principal metadata: got %+v", principal)
	}

	// 3. List keys for project
	keys, err := svc.ListProjectAPIKeys(ctx, projectID)
	if err != nil {
		t.Fatalf("ListProjectAPIKeys failed: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}

	// 4. Revoke API Key
	err = svc.RevokeAPIKey(ctx, created.ID)
	if err != nil {
		t.Fatalf("RevokeAPIKey failed: %v", err)
	}

	// 5. Authenticate after revocation should fail
	_, err = svc.AuthenticateAPIKey(ctx, created.SecretKey)
	if err == nil {
		t.Fatalf("expected error authenticating with revoked key, got nil")
	}
}

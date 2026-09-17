package provider

import (
	"context"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/provider/dto"
	"github.com/google/uuid"
)

type mockCredentialRepo struct {
	store map[string]*ProviderCredential
}

func newMockRepo() *mockCredentialRepo {
	return &mockCredentialRepo{store: make(map[string]*ProviderCredential)}
}

func key(orgID uuid.UUID, provider string) string {
	return orgID.String() + ":" + provider
}

func (m *mockCredentialRepo) Save(ctx context.Context, cred *ProviderCredential) error {
	m.store[key(cred.OrganizationID, cred.Provider)] = cred
	return nil
}

func (m *mockCredentialRepo) GetByOrgAndProvider(ctx context.Context, orgID uuid.UUID, provider string) (*ProviderCredential, error) {
	c, ok := m.store[key(orgID, provider)]
	if !ok {
		return nil, ErrCredentialNotFound
	}
	return c, nil
}

func (m *mockCredentialRepo) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*ProviderCredential, error) {
	var list []*ProviderCredential
	for _, c := range m.store {
		if c.OrganizationID == orgID {
			list = append(list, c)
		}
	}
	return list, nil
}

func (m *mockCredentialRepo) Delete(ctx context.Context, orgID uuid.UUID, provider string) error {
	k := key(orgID, provider)
	if _, ok := m.store[k]; !ok {
		return ErrCredentialNotFound
	}
	delete(m.store, k)
	return nil
}

func TestProviderCredentialService(t *testing.T) {
	repo := newMockRepo()
	secret := "test-secret-key-32-bytes-long!"
	svc := NewProviderCredentialService(repo, secret)

	orgID := uuid.New()

	// 1. Test Save Credential
	req := dto.SaveCredentialRequest{
		Provider: "openai",
		APIKey:   "sk-proj-1234567890abcdef",
	}

	resp, err := svc.SaveCredential(context.Background(), orgID, req)
	if err != nil {
		t.Fatalf("SaveCredential failed: %v", err)
	}

	if resp.Provider != "openai" {
		t.Errorf("Expected provider 'openai', got '%s'", resp.Provider)
	}
	if resp.MaskedAPIKey != "sk-****cdef" {
		t.Errorf("Expected masked key 'sk-****cdef', got '%s'", resp.MaskedAPIKey)
	}

	// 2. Test Get Decrypted Key
	decryptedKey, _, err := svc.GetDecryptedKey(context.Background(), orgID, "openai")
	if err != nil {
		t.Fatalf("GetDecryptedKey failed: %v", err)
	}
	if decryptedKey != "sk-proj-1234567890abcdef" {
		t.Errorf("Expected decrypted key 'sk-proj-1234567890abcdef', got '%s'", decryptedKey)
	}

	// 3. Test List Credentials
	list, err := svc.ListCredentials(context.Background(), orgID)
	if err != nil {
		t.Fatalf("ListCredentials failed: %v", err)
	}
	if len(list.Data) != 1 {
		t.Errorf("Expected 1 credential, got %d", len(list.Data))
	}

	// 4. Test Delete Credential
	if err := svc.DeleteCredential(context.Background(), orgID, "openai"); err != nil {
		t.Fatalf("DeleteCredential failed: %v", err)
	}

	_, _, err = svc.GetDecryptedKey(context.Background(), orgID, "openai")
	if err == nil {
		t.Errorf("Expected error after delete, got nil")
	}
}

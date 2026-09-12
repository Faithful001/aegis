package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/auth"
	authDto "github.com/Faithful001/aegis/internal/domain/auth/dto"
	"github.com/Faithful001/aegis/internal/infra/middleware"
	"github.com/gin-gonic/gin"
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
	k, ok := m.keys[id]
	if !ok {
		return nil, auth.ErrAPIKeyNotFound
	}
	return k, nil
}
func (m *mockAPIKeyRepo) GetByHash(ctx context.Context, hash string) (*auth.APIKey, error) {
	k, ok := m.byHash[hash]
	if !ok {
		return nil, auth.ErrAPIKeyNotFound
	}
	return k, nil
}
func (m *mockAPIKeyRepo) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*auth.APIKey, error) {
	return nil, nil
}
func (m *mockAPIKeyRepo) Update(ctx context.Context, key *auth.APIKey) error { return nil }
func (m *mockAPIKeyRepo) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockAPIKeyRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	if k, ok := m.keys[id]; ok {
		k.Status = auth.APIKeyStatusRevoked
	}
	return nil
}
func (m *mockAPIKeyRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

func setupTestEngine(apiKeyService *auth.APIKeyService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.APIKeyAuthMiddleware(apiKeyService))
	r.GET("/v1/test", func(c *gin.Context) {
		principal, ok := middleware.GetPrincipal(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "principal not found in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":          "authenticated",
			"organization_id": principal.OrganizationID.String(),
			"project_id":      principal.ProjectID.String(),
		})
	})
	return r
}

func TestAPIKeyAuthMiddleware(t *testing.T) {
	repo := newMockAPIKeyRepo()
	generator := auth.NewAPIKeyGenerator()
	apiKeyService := auth.NewAPIKeyService(repo, generator)
	engine := setupTestEngine(apiKeyService)

	orgID := uuid.New()
	projectID := uuid.New()
	userID := uuid.New()

	createdKey, err := apiKeyService.CreateAPIKey(context.Background(), orgID, projectID, userID, authDto.CreateAPIKeyRequest{
		Name: "Test Key",
	})
	if err != nil {
		t.Fatalf("failed to create api key: %v", err)
	}

	// 1. Success with Authorization: Bearer <key>
	req, _ := http.NewRequest("GET", "/v1/test", nil)
	req.Header.Set("Authorization", "Bearer "+createdKey.SecretKey)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// 2. Success with X-API-Key: <key>
	req2, _ := http.NewRequest("GET", "/v1/test", nil)
	req2.Header.Set("X-API-Key", createdKey.SecretKey)
	w2 := httptest.NewRecorder()
	engine.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200 with X-API-Key, got %d", w2.Code)
	}

	// 3. Missing API key (401)
	req3, _ := http.NewRequest("GET", "/v1/test", nil)
	w3 := httptest.NewRecorder()
	engine.ServeHTTP(w3, req3)
	if w3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing key, got %d", w3.Code)
	}

	// 4. Invalid API key (401)
	req4, _ := http.NewRequest("GET", "/v1/test", nil)
	req4.Header.Set("Authorization", "Bearer aeg_live_invalidkey1234567890")
	w4 := httptest.NewRecorder()
	engine.ServeHTTP(w4, req4)
	if w4.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid key, got %d", w4.Code)
	}

	// 5. Revoked API key (401)
	_ = apiKeyService.RevokeAPIKey(context.Background(), createdKey.ID)
	req5, _ := http.NewRequest("GET", "/v1/test", nil)
	req5.Header.Set("Authorization", "Bearer "+createdKey.SecretKey)
	w5 := httptest.NewRecorder()
	engine.ServeHTTP(w5, req5)
	if w5.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for revoked key, got %d", w5.Code)
	}
}

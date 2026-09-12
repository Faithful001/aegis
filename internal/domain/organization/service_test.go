package organization_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/organization"
	orgDto "github.com/Faithful001/aegis/internal/domain/organization/dto"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	users map[uuid.UUID]*user.User
}

func (m *mockUserRepo) Create(ctx context.Context, u *user.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, u *user.User) error { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error       { return nil }

type mockOrgRepo struct {
	orgs    map[uuid.UUID]*organization.Organization
	slugs   map[string]*organization.Organization
	members map[string]*organization.OrganizationMember
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{
		orgs:    make(map[uuid.UUID]*organization.Organization),
		slugs:   make(map[string]*organization.Organization),
		members: make(map[string]*organization.OrganizationMember),
	}
}

func (m *mockOrgRepo) Create(ctx context.Context, org *organization.Organization, ownerID uuid.UUID) error {
	m.orgs[org.ID] = org
	m.slugs[org.Slug] = org
	m.members[org.ID.String()+":"+ownerID.String()] = &organization.OrganizationMember{
		OrganizationID: org.ID,
		UserID:         ownerID,
		Role:           organization.RoleOwner,
	}
	return nil
}
func (m *mockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	org, ok := m.orgs[id]
	if !ok {
		return nil, organization.ErrOrgNotFound
	}
	return org, nil
}
func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	org, ok := m.slugs[slug]
	if !ok {
		return nil, organization.ErrOrgNotFound
	}
	return org, nil
}
func (m *mockOrgRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*organization.Organization, error) {
	var list []*organization.Organization
	for _, member := range m.members {
		if member.UserID == userID {
			if o, ok := m.orgs[member.OrganizationID]; ok {
				list = append(list, o)
			}
		}
	}
	return list, nil
}
func (m *mockOrgRepo) Update(ctx context.Context, org *organization.Organization) error {
	m.orgs[org.ID] = org
	return nil
}
func (m *mockOrgRepo) AddMember(ctx context.Context, member *organization.OrganizationMember) error {
	m.members[member.OrganizationID.String()+":"+member.UserID.String()] = member
	return nil
}
func (m *mockOrgRepo) GetMember(ctx context.Context, orgID, userID uuid.UUID) (*organization.OrganizationMember, error) {
	mem, ok := m.members[orgID.String()+":"+userID.String()]
	if !ok {
		return nil, organization.ErrMemberNotFound
	}
	return mem, nil
}
func (m *mockOrgRepo) ListMembers(ctx context.Context, orgID uuid.UUID) ([]*organization.OrganizationMember, error) {
	var list []*organization.OrganizationMember
	for _, mem := range m.members {
		if mem.OrganizationID == orgID {
			list = append(list, mem)
		}
	}
	return list, nil
}
func (m *mockOrgRepo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	delete(m.members, orgID.String()+":"+userID.String())
	return nil
}

func TestOrganizationService_Lifecycle(t *testing.T) {
	userID := uuid.New()
	userRepo := &mockUserRepo{users: map[uuid.UUID]*user.User{userID: {ID: userID, Email: "owner@example.com"}}}
	orgRepo := newMockOrgRepo()

	svc := organization.NewService(orgRepo, userRepo)
	ctx := context.Background()

	// 1. Create Organization
	req := orgDto.CreateOrganizationRequest{
		Name: "Antigravity AI Inc",
	}
	orgResp, err := svc.CreateOrganization(ctx, userID, req)
	if err != nil {
		t.Fatalf("failed to create organization: %v", err)
	}

	if orgResp.Role != "owner" {
		t.Fatalf("expected role 'owner', got %s", orgResp.Role)
	}

	// 2. Get Organization
	fetched, err := svc.GetOrganization(ctx, orgResp.ID, userID)
	if err != nil {
		t.Fatalf("failed to get organization: %v", err)
	}
	if fetched.Name != "Antigravity AI Inc" {
		t.Errorf("expected name 'Antigravity AI Inc', got %s", fetched.Name)
	}

	// 3. Unauthorized access by outsider
	outsiderID := uuid.New()
	_, err = svc.GetOrganization(ctx, orgResp.ID, outsiderID)
	if !errors.Is(err, organization.ErrUnauthorizedTenant) {
		t.Errorf("expected ErrUnauthorizedTenant, got %v", err)
	}
}

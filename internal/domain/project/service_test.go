package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Faithful001/aegis/internal/domain/organization"
	"github.com/Faithful001/aegis/internal/domain/project"
	projectDto "github.com/Faithful001/aegis/internal/domain/project/dto"
	"github.com/google/uuid"
)

type mockProjectRepo struct {
	projects map[uuid.UUID]*project.Project
	bySlug   map[string]*project.Project
}

func newMockProjectRepo() *mockProjectRepo {
	return &mockProjectRepo{
		projects: make(map[uuid.UUID]*project.Project),
		bySlug:   make(map[string]*project.Project),
	}
}

func (m *mockProjectRepo) Create(ctx context.Context, p *project.Project) error {
	m.projects[p.ID] = p
	m.bySlug[p.OrganizationID.String()+":"+p.Slug] = p
	return nil
}
func (m *mockProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, project.ErrProjectNotFound
	}
	return p, nil
}
func (m *mockProjectRepo) GetByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug string) (*project.Project, error) {
	p, ok := m.bySlug[orgID.String()+":"+slug]
	if !ok {
		return nil, project.ErrProjectNotFound
	}
	return p, nil
}
func (m *mockProjectRepo) ListByOrganization(ctx context.Context, orgID uuid.UUID) ([]*project.Project, error) {
	var list []*project.Project
	for _, p := range m.projects {
		if p.OrganizationID == orgID {
			list = append(list, p)
		}
	}
	return list, nil
}
func (m *mockProjectRepo) Update(ctx context.Context, p *project.Project) error {
	m.projects[p.ID] = p
	return nil
}
func (m *mockProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	delete(m.projects, id)
	return nil
}

type mockOrgRepo struct {
	members map[string]*organization.OrganizationMember
}

func newMockOrgRepo() *mockOrgRepo {
	return &mockOrgRepo{members: make(map[string]*organization.OrganizationMember)}
}

func (m *mockOrgRepo) Create(ctx context.Context, org *organization.Organization, ownerID uuid.UUID) error {
	return nil
}
func (m *mockOrgRepo) GetByID(ctx context.Context, id uuid.UUID) (*organization.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) GetBySlug(ctx context.Context, slug string) (*organization.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*organization.Organization, error) {
	return nil, nil
}
func (m *mockOrgRepo) Update(ctx context.Context, org *organization.Organization) error {
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
	return nil, nil
}
func (m *mockOrgRepo) RemoveMember(ctx context.Context, orgID, userID uuid.UUID) error {
	return nil
}

func TestProjectService_Lifecycle(t *testing.T) {
	projectRepo := newMockProjectRepo()
	orgRepo := newMockOrgRepo()

	svc := project.NewProjectService(projectRepo, orgRepo)
	ctx := context.Background()

	orgID := uuid.New()
	ownerID := uuid.New()
	strangerID := uuid.New()

	_ = orgRepo.AddMember(ctx, &organization.OrganizationMember{
		OrganizationID: orgID,
		UserID:         ownerID,
		Role:           organization.RoleOwner,
	})

	// 1. Create Project
	req := projectDto.CreateProjectRequest{
		Name:           "Core Inference API",
		MaxConcurrency: 25,
		RateLimitRPM:   1200,
		RateLimitTPM:   200000,
	}

	projResp, err := svc.CreateProject(ctx, orgID, ownerID, req)
	if err != nil {
		t.Fatalf("CreateProject failed: %v", err)
	}

	if projResp.MaxConcurrency != 25 || projResp.RateLimitRPM != 1200 {
		t.Errorf("custom quotas not respected: %+v", projResp)
	}

	// 2. Stranger cannot create project in this org
	_, err = svc.CreateProject(ctx, orgID, strangerID, req)
	if !errors.Is(err, organization.ErrUnauthorizedTenant) {
		t.Errorf("expected ErrUnauthorizedTenant, got %v", err)
	}

	// 3. Get Project
	fetched, err := svc.GetProject(ctx, projResp.ID, ownerID)
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}
	if fetched.Name != "Core Inference API" {
		t.Errorf("expected name 'Core Inference API', got %s", fetched.Name)
	}
}

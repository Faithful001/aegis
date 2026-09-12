package project

import (
	"context"
	"fmt"

	"github.com/Faithful001/aegis/internal/domain/organization"
	"github.com/Faithful001/aegis/internal/domain/project/dto"
	"github.com/google/uuid"
)

type Service struct {
	projectRepo Repository
	orgRepo     organization.Repository
}

func NewService(projectRepo Repository, orgRepo organization.Repository) *Service {
	return &Service{
		projectRepo: projectRepo,
		orgRepo:     orgRepo,
	}
}

func (s *Service) CreateProject(ctx context.Context, orgID, userID uuid.UUID, req dto.CreateProjectRequest) (*dto.ProjectResponse, error) {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil || member == nil {
		return nil, organization.ErrUnauthorizedTenant
	}
	if member.Role != organization.RoleOwner && member.Role != organization.RoleAdmin {
		return nil, organization.ErrUnauthorizedTenant
	}

	newProject, err := NewProject(orgID, req.Name, req.Slug)
	if err != nil {
		return nil, err
	}

	if req.MaxConcurrency > 0 {
		newProject.MaxConcurrency = req.MaxConcurrency
	}
	if req.RateLimitRPM > 0 {
		newProject.RateLimitRPM = req.RateLimitRPM
	}
	if req.RateLimitTPM > 0 {
		newProject.RateLimitTPM = req.RateLimitTPM
	}

	existing, err := s.projectRepo.GetByOrgAndSlug(ctx, orgID, newProject.Slug)
	if err == nil && existing != nil {
		return nil, ErrProjectSlugExists
	}

	if err := s.projectRepo.Create(ctx, newProject); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return toProjectResponse(newProject), nil
}

func (s *Service) GetProject(ctx context.Context, projectID, userID uuid.UUID) (*dto.ProjectResponse, error) {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	member, err := s.orgRepo.GetMember(ctx, proj.OrganizationID, userID)
	if err != nil || member == nil {
		return nil, organization.ErrUnauthorizedTenant
	}

	return toProjectResponse(proj), nil
}

func (s *Service) ListOrganizationProjects(ctx context.Context, orgID, userID uuid.UUID) ([]*dto.ProjectResponse, error) {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil || member == nil {
		return nil, organization.ErrUnauthorizedTenant
	}

	projects, err := s.projectRepo.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.ProjectResponse, len(projects))
	for i, p := range projects {
		responses[i] = toProjectResponse(p)
	}
	return responses, nil
}

func (s *Service) DeleteProject(ctx context.Context, projectID, userID uuid.UUID) error {
	proj, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	member, err := s.orgRepo.GetMember(ctx, proj.OrganizationID, userID)
	if err != nil || member == nil {
		return organization.ErrUnauthorizedTenant
	}
	if member.Role != organization.RoleOwner && member.Role != organization.RoleAdmin {
		return organization.ErrUnauthorizedTenant
	}

	return s.projectRepo.Delete(ctx, projectID)
}

func toProjectResponse(p *Project) *dto.ProjectResponse {
	return &dto.ProjectResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		Name:           p.Name,
		Slug:           p.Slug,
		Status:         string(p.Status),
		MaxConcurrency: p.MaxConcurrency,
		RateLimitRPM:   p.RateLimitRPM,
		RateLimitTPM:   p.RateLimitTPM,
		CreatedAt:      p.CreatedAt,
	}
}

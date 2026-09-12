package organization

import (
	"context"
	"fmt"

	"github.com/Faithful001/aegis/internal/domain/organization/dto"
	"github.com/Faithful001/aegis/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	orgRepo  Repository
	userRepo user.Repository
}

func NewService(orgRepo Repository, userRepo user.Repository) *Service {
	return &Service{
		orgRepo:  orgRepo,
		userRepo: userRepo,
	}
}

func (s *Service) CreateOrganization(ctx context.Context, userID uuid.UUID, req dto.CreateOrganizationRequest) (*dto.OrganizationResponse, error) {
	if _, err := s.userRepo.GetByID(ctx, userID); err != nil {
		return nil, user.ErrUserNotFound
	}

	newOrg, err := NewOrganization(req.Name, req.Slug)
	if err != nil {
		return nil, err
	}

	existing, err := s.orgRepo.GetBySlug(ctx, newOrg.Slug)
	if err == nil && existing != nil {
		return nil, ErrOrgSlugExists
	}

	if err := s.orgRepo.Create(ctx, newOrg, userID); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	return &dto.OrganizationResponse{
		ID:        newOrg.ID,
		Name:      newOrg.Name,
		Slug:      newOrg.Slug,
		Status:    string(newOrg.Status),
		Role:      string(RoleOwner),
		CreatedAt: newOrg.CreatedAt,
	}, nil
}

func (s *Service) GetOrganization(ctx context.Context, orgID, userID uuid.UUID) (*dto.OrganizationResponse, error) {
	member, err := s.orgRepo.GetMember(ctx, orgID, userID)
	if err != nil || member == nil {
		return nil, ErrUnauthorizedTenant
	}

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &dto.OrganizationResponse{
		ID:        org.ID,
		Name:      org.Name,
		Slug:      org.Slug,
		Status:    string(org.Status),
		Role:      string(member.Role),
		CreatedAt: org.CreatedAt,
	}, nil
}

func (s *Service) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*dto.OrganizationResponse, error) {
	orgs, err := s.orgRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.OrganizationResponse, len(orgs))
	for i, o := range orgs {
		role := ""
		if m, err := s.orgRepo.GetMember(ctx, o.ID, userID); err == nil && m != nil {
			role = string(m.Role)
		}
		responses[i] = &dto.OrganizationResponse{
			ID:        o.ID,
			Name:      o.Name,
			Slug:      o.Slug,
			Status:    string(o.Status),
			Role:      role,
			CreatedAt: o.CreatedAt,
		}
	}
	return responses, nil
}

func (s *Service) AddMember(ctx context.Context, orgID, currentUserID uuid.UUID, req dto.AddMemberRequest) (*dto.MemberResponse, error) {
	callerMember, err := s.orgRepo.GetMember(ctx, orgID, currentUserID)
	if err != nil || callerMember == nil {
		return nil, ErrUnauthorizedTenant
	}
	if callerMember.Role != RoleOwner && callerMember.Role != RoleAdmin {
		return nil, ErrUnauthorizedTenant
	}

	if _, err := s.userRepo.GetByID(ctx, req.UserID); err != nil {
		return nil, user.ErrUserNotFound
	}

	existing, _ := s.orgRepo.GetMember(ctx, orgID, req.UserID)
	if existing != nil {
		return nil, ErrMemberAlreadyExists
	}

	member := &OrganizationMember{
		ID:             uuid.New(),
		OrganizationID: orgID,
		UserID:         req.UserID,
		Role:           OrgRole(req.Role),
	}

	if err := s.orgRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	return &dto.MemberResponse{
		ID:        member.ID,
		UserID:    member.UserID,
		Role:      string(member.Role),
		CreatedAt: member.CreatedAt,
	}, nil
}

func (s *Service) ListMembers(ctx context.Context, orgID, currentUserID uuid.UUID) ([]*dto.MemberResponse, error) {
	member, err := s.orgRepo.GetMember(ctx, orgID, currentUserID)
	if err != nil || member == nil {
		return nil, ErrUnauthorizedTenant
	}

	members, err := s.orgRepo.ListMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.MemberResponse, len(members))
	for i, m := range members {
		responses[i] = &dto.MemberResponse{
			ID:        m.ID,
			UserID:    m.UserID,
			Role:      string(m.Role),
			CreatedAt: m.CreatedAt,
		}
	}
	return responses, nil
}

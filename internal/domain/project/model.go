package project

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectStatus string

const (
	StatusActive    ProjectStatus = "active"
	StatusSuspended ProjectStatus = "suspended"
)

type Project struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID     `gorm:"type:uuid;not null;index:idx_org_project_slug,unique;index" json:"organization_id"`
	Name           string        `gorm:"type:varchar(255);not null" json:"name"`
	Slug           string        `gorm:"type:varchar(100);not null;index:idx_org_project_slug,unique" json:"slug"`
	Status         ProjectStatus `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	MaxConcurrency int           `gorm:"not null;default:10" json:"max_concurrency"`
	RateLimitRPM   int           `gorm:"not null;default:600" json:"rate_limit_rpm"`   // Requests per minute
	RateLimitTPM   int           `gorm:"not null;default:100000" json:"rate_limit_tpm"` // Tokens per minute
	CreatedAt      time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Project) TableName() string {
	return "projects"
}

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Status == "" {
		p.Status = StatusActive
	}
	if p.MaxConcurrency <= 0 {
		p.MaxConcurrency = 10
	}
	if p.RateLimitRPM <= 0 {
		p.RateLimitRPM = 600
	}
	if p.RateLimitTPM <= 0 {
		p.RateLimitTPM = 100000
	}
	return nil
}

var projectSlugRegex = regexp.MustCompile(`[^a-z0-9\-]+`)

func GenerateProjectSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = projectSlugRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "proj-" + uuid.New().String()[:8]
	}
	return s
}

func NewProject(orgID uuid.UUID, name, slug string) (*Project, error) {
	if orgID == uuid.Nil {
		return nil, ErrInvalidProjectData
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidProjectData
	}
	if slug == "" {
		slug = GenerateProjectSlug(name)
	} else {
		slug = GenerateProjectSlug(slug)
	}

	return &Project{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Name:           name,
		Slug:           slug,
		Status:         StatusActive,
		MaxConcurrency: 10,
		RateLimitRPM:   600,
		RateLimitTPM:   100000,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}, nil
}

func (p *Project) IsActive() bool {
	return p.Status == StatusActive
}

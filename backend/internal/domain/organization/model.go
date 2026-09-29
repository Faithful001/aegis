package organization

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrgRole string

const (
	RoleOwner  OrgRole = "owner"
	RoleAdmin  OrgRole = "admin"
	RoleMember OrgRole = "member"
)

type OrgStatus string

const (
	StatusActive    OrgStatus = "active"
	StatusSuspended OrgStatus = "suspended"
)

type Organization struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name      string       `gorm:"type:varchar(255);not null" json:"name"`
	Slug      string       `gorm:"type:varchar(100);not null;uniqueIndex" json:"slug"`
	Status    OrgStatus    `gorm:"type:varchar(50);not null;default:'active'" json:"status"`
	Members   []OrganizationMember `gorm:"foreignKey:OrganizationID" json:"members,omitempty"`
	CreatedAt time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
}

func (Organization) TableName() string {
	return "organizations"
}

func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.Status == "" {
		o.Status = StatusActive
	}
	return nil
}

type OrganizationMember struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OrganizationID uuid.UUID `gorm:"type:uuid;not null;index:idx_org_user,unique" json:"organization_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;index:idx_org_user,unique" json:"user_id"`
	Role           OrgRole   `gorm:"type:varchar(50);not null;default:'member'" json:"role"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (OrganizationMember) TableName() string {
	return "organization_members"
}

func (m *OrganizationMember) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Role == "" {
		m.Role = RoleMember
	}
	return nil
}

var slugRegex = regexp.MustCompile(`[^a-z0-9\-]+`)

func GenerateSlug(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "org-" + uuid.New().String()[:8]
	}
	return s
}

func NewOrganization(name, slug string) (*Organization, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidOrgData
	}
	if slug == "" {
		slug = GenerateSlug(name)
	} else {
		slug = GenerateSlug(slug)
	}

	return &Organization{
		ID:        uuid.New(),
		Name:      name,
		Slug:      slug,
		Status:    StatusActive,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}

func (o *Organization) IsActive() bool {
	return o.Status == StatusActive
}

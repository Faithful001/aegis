package project_test

import (
	"testing"

	"github.com/Faithful001/aegis/internal/domain/project"
	"github.com/google/uuid"
)

func TestNewProject_Success(t *testing.T) {
	orgID := uuid.New()
	proj, err := project.NewProject(orgID, "Production LLM Gateway", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if proj.OrganizationID != orgID {
		t.Errorf("expected org ID %v, got %v", orgID, proj.OrganizationID)
	}

	if proj.Slug != "production-llm-gateway" {
		t.Errorf("expected slug 'production-llm-gateway', got %s", proj.Slug)
	}

	if proj.MaxConcurrency != 10 {
		t.Errorf("expected default max concurrency 10, got %d", proj.MaxConcurrency)
	}

	if proj.RateLimitRPM != 600 {
		t.Errorf("expected default RPM 600, got %d", proj.RateLimitRPM)
	}
}

func TestNewProject_Invalid(t *testing.T) {
	_, err := project.NewProject(uuid.Nil, "Name", "")
	if err == nil {
		t.Errorf("expected error for nil orgID, got nil")
	}

	_, err = project.NewProject(uuid.New(), "", "")
	if err == nil {
		t.Errorf("expected error for empty name, got nil")
	}
}

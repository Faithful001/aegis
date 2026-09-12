package organization_test

import (
	"testing"

	"github.com/Faithful001/aegis/internal/domain/organization"
)

func TestNewOrganization_Success(t *testing.T) {
	org, err := organization.NewOrganization("Acme AI Labs", "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if org.Name != "Acme AI Labs" {
		t.Errorf("expected name 'Acme AI Labs', got %s", org.Name)
	}

	if org.Slug != "acme-ai-labs" {
		t.Errorf("expected generated slug 'acme-ai-labs', got %s", org.Slug)
	}

	if !org.IsActive() {
		t.Errorf("expected organization to be active")
	}
}

func TestNewOrganization_CustomSlug(t *testing.T) {
	org, err := organization.NewOrganization("Acme AI Labs", "custom-slug-123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if org.Slug != "custom-slug-123" {
		t.Errorf("expected custom slug, got %s", org.Slug)
	}
}

func TestNewOrganization_InvalidName(t *testing.T) {
	_, err := organization.NewOrganization("", "")
	if err == nil {
		t.Errorf("expected error for empty name, got nil")
	}
}

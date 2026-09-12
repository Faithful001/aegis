package organization

import "errors"

var (
	ErrOrgNotFound            = errors.New("organization not found")
	ErrOrgSlugExists          = errors.New("organization slug already exists")
	ErrInvalidOrgData         = errors.New("invalid organization data")
	ErrMemberAlreadyExists    = errors.New("user is already a member of this organization")
	ErrMemberNotFound         = errors.New("member not found in organization")
	ErrUnauthorizedTenant     = errors.New("unauthorized access to organization")
	ErrCannotRemoveLastOwner  = errors.New("cannot remove the last owner of an organization")
)

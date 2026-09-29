package project

import "errors"

var (
	ErrProjectNotFound       = errors.New("project not found")
	ErrProjectSlugExists     = errors.New("project slug already exists in this organization")
	ErrInvalidProjectData    = errors.New("invalid project data")
	ErrProjectSuspended      = errors.New("project is suspended or inactive")
	ErrProjectQuotaExceeded  = errors.New("project quota exceeded")
)

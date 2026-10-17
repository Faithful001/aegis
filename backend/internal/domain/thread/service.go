package thread

import (
	"context"
	"time"

	"github.com/Faithful001/aegis/internal/domain/thread/dto"
	"github.com/google/uuid"
)

type ThreadService struct {
	repo IThreadRepository
}

func NewThreadService(repo IThreadRepository) *ThreadService {
	return &ThreadService{repo: repo}
}

func (s *ThreadService) CreateThread(ctx context.Context, userID uuid.UUID, req dto.CreateThreadRequest) (*dto.ThreadResponse, error) {
	title := req.Title
	if title == "" {
		title = "New Exploration"
	}

	t := &Thread{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     title,
		Model:     req.Model,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateThread(ctx, t); err != nil {
		return nil, err
	}

	resp := mapThreadToResponse(t)
	return &resp, nil
}

func (s *ThreadService) GetThread(ctx context.Context, userID, threadID uuid.UUID) (*dto.ThreadResponse, error) {
	t, err := s.repo.GetThreadByID(ctx, userID, threadID)
	if err != nil {
		return nil, err
	}

	resp := mapThreadToResponse(t)
	return &resp, nil
}

func (s *ThreadService) ListThreads(ctx context.Context, userID uuid.UUID) ([]dto.ThreadResponse, error) {
	threads, err := s.repo.ListThreadsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.ThreadResponse, len(threads))
	for i, t := range threads {
		responses[i] = mapThreadToResponse(t)
	}
	return responses, nil
}

func (s *ThreadService) UpdateThread(ctx context.Context, userID, threadID uuid.UUID, req dto.UpdateThreadRequest) (*dto.ThreadResponse, error) {
	t, err := s.repo.GetThreadByID(ctx, userID, threadID)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Model != nil {
		t.Model = *req.Model
	}
	t.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateThread(ctx, t); err != nil {
		return nil, err
	}

	resp := mapThreadToResponse(t)
	return &resp, nil
}

func (s *ThreadService) DeleteThread(ctx context.Context, userID, threadID uuid.UUID) error {
	return s.repo.DeleteThread(ctx, userID, threadID)
}

func (s *ThreadService) AddMessage(ctx context.Context, userID, threadID uuid.UUID, req dto.CreateMessageRequest) (*dto.MessageResponse, error) {
	t, err := s.repo.GetThreadByID(ctx, userID, threadID)
	if err != nil {
		return nil, err
	}

	msg := &Message{
		ID:        uuid.New(),
		ThreadID:  t.ID,
		Role:      req.Role,
		Content:   req.Content,
		Model:     req.Model,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.CreateMessage(ctx, msg); err != nil {
		return nil, err
	}

	// Update thread updated_at timestamp and title if it's the first user message
	if t.Title == "New Exploration" && req.Role == "user" && len(req.Content) > 0 {
		title := req.Content
		if len(title) > 40 {
			title = title[:40] + "..."
		}
		t.Title = title
	}
	t.UpdatedAt = time.Now().UTC()
	_ = s.repo.UpdateThread(ctx, t)

	resp := mapMessageToResponse(msg)
	return &resp, nil
}

func (s *ThreadService) ListMessages(ctx context.Context, userID, threadID uuid.UUID) ([]dto.MessageResponse, error) {
	_, err := s.repo.GetThreadByID(ctx, userID, threadID)
	if err != nil {
		return nil, err
	}

	msgs, err := s.repo.ListMessagesByThread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.MessageResponse, len(msgs))
	for i, m := range msgs {
		res[i] = mapMessageToResponse(m)
	}
	return res, nil
}

func mapThreadToResponse(t *Thread) dto.ThreadResponse {
	var msgs []dto.MessageResponse
	if len(t.Messages) > 0 {
		msgs = make([]dto.MessageResponse, len(t.Messages))
		for i, m := range t.Messages {
			msgs[i] = mapMessageToResponse(&m)
		}
	} else {
		msgs = []dto.MessageResponse{}
	}

	return dto.ThreadResponse{
		ID:        t.ID,
		UserID:    t.UserID,
		Title:     t.Title,
		Model:     t.Model,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
		Messages:  msgs,
	}
}

func mapMessageToResponse(m *Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID:        m.ID,
		ThreadID:  m.ThreadID,
		Role:      m.Role,
		Content:   m.Content,
		Model:     m.Model,
		CreatedAt: m.CreatedAt,
	}
}

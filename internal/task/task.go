package task

import (
	"time"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type Task struct {
	ID          string     `json:"id"`
	Content     string     `json:"content"`
	Priority    Priority   `json:"priority"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	Tags        []string   `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Archived    bool       `json:"archived"`
}

func New(content string) *Task {
	now := time.Now()
	return &Task{
		ID:        uuid.New().String(),
		Content:   content,
		Priority:  PriorityMedium,
		Tags:      []string{},
		CreatedAt: now,
		UpdatedAt: now,
		Archived:  false,
	}
}

func (t *Task) Complete() {
	now := time.Now()
	t.CompletedAt = &now
	t.UpdatedAt = now
}

func (t *Task) Uncomplete() {
	t.CompletedAt = nil
	t.UpdatedAt = time.Now()
}

func (t *Task) Archive() {
	t.Archived = true
	t.UpdatedAt = time.Now()
}

func (t *Task) Unarchive() {
	t.Archived = false
	t.UpdatedAt = time.Now()
}

func (t *Task) SetPriority(p Priority) {
	t.Priority = p
	t.UpdatedAt = time.Now()
}

func (t *Task) SetDeadline(d *time.Time) {
	t.Deadline = d
	t.UpdatedAt = time.Now()
}

func (t *Task) IsOverdue() bool {
	if t.Deadline == nil || t.CompletedAt != nil {
		return false
	}
	return time.Now().After(*t.Deadline)
}

func (t *Task) CyclePriority() {
	switch t.Priority {
	case PriorityHigh:
		t.Priority = PriorityMedium
	case PriorityMedium:
		t.Priority = PriorityLow
	case PriorityLow:
		t.Priority = PriorityHigh
	}
	t.UpdatedAt = time.Now()
}

package training

import "time"

type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
)

type Training struct {
    ID        int       `json:"id"`
    ModelID   int       `json:"model_id"`
    Text      string    `json:"-"`
    Status    Status    `json:"status"`
    Error     string    `json:"error,omitempty"`
    CreatedAt time.Time `json:"created_at"`
}

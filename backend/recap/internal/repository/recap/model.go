package recap

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type recapModel struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Year                 int32
	Archetype            string
	ArchetypeTitle       string
	ArchetypeDescription string
	ArchetypeReasons     []archetypeReasonModel
	Slides               json.RawMessage
	GeneratedAt          time.Time
}

type archetypeReasonModel struct {
	Metric      string `json:"metric"`
	Value       string `json:"value"`
	Explanation string `json:"explanation"`
}

func (m *recapModel) Scan(row rowScanner) error {
	return row.Scan(
		&m.ID,
		&m.UserID,
		&m.Year,
		&m.Archetype,
		&m.ArchetypeTitle,
		&m.ArchetypeDescription,
		&m.ArchetypeReasons,
		&m.Slides,
		&m.GeneratedAt,
	)
}

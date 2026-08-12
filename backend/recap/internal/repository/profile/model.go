package profile

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type rowScanner interface {
	Scan(dest ...any) error
}

type profileModel struct {
	ID           uuid.UUID
	Name         string
	Surname      string
	AvatarURL    pgtype.Text
	Hint         pgtype.Text
	RegisteredAt time.Time
}

func (m *profileModel) Scan(row rowScanner) error {
	return row.Scan(
		&m.ID,
		&m.Name,
		&m.Surname,
		&m.AvatarURL,
		&m.Hint,
		&m.RegisteredAt,
	)
}

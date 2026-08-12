package profile

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (entity.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		SELECT
			id,
			name,
			surname,
			avatar_url,
			hint,
			created_at
		FROM recap.users
		WHERE id = $1
	`

	var model profileModel
	err := model.Scan(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Profile{}, fmt.Errorf(
			"get profile %s: %w",
			id,
			entity.ErrProfileNotFound,
		)
	}
	if err != nil {
		return entity.Profile{}, fmt.Errorf("get profile by id: %w", err)
	}

	return profileModelToEntity(model), nil
}

package recap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) GetByProfileAndYear(
	ctx context.Context,
	profileID uuid.UUID,
	year int32,
) (entity.Recap, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		SELECT
			id,
			user_id,
			year,
			archetype,
			archetype_title,
			archetype_description,
			archetype_reasons,
			slides,
			generated_at
		FROM recap.recaps
		WHERE user_id = $1
		  AND year = $2
	`

	var model recapModel
	err := model.Scan(r.pool.QueryRow(ctx, query, profileID, year))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Recap{}, fmt.Errorf(
			"get recap for profile %s and year %d: %w",
			profileID,
			year,
			entity.ErrRecapNotFound,
		)
	}
	if err != nil {
		return entity.Recap{}, fmt.Errorf(
			"get recap by profile and year: %w",
			err,
		)
	}

	recap, err := recapModelToEntity(model)
	if err != nil {
		return entity.Recap{}, fmt.Errorf("convert recap model: %w", err)
	}

	return recap, nil
}

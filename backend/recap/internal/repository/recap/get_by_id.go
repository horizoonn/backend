package recap

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
		WHERE id = $1
	`

	var model recapModel
	err := model.Scan(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.Recap{}, fmt.Errorf("get recap %s: %w", id, entity.ErrRecapNotFound)
	}
	if err != nil {
		return entity.Recap{}, fmt.Errorf("get recap by id: %w", err)
	}

	recap, err := recapModelToEntity(model)
	if err != nil {
		return entity.Recap{}, fmt.Errorf("convert recap model: %w", err)
	}

	return recap, nil
}

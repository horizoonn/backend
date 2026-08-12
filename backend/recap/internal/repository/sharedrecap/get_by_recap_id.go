package sharedrecap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) GetByRecapID(
	ctx context.Context,
	recapID uuid.UUID,
) (entity.SharedRecap, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		SELECT
			token,
			recap_id,
			snapshot,
			created_at
		FROM recap.shared_recaps
		WHERE recap_id = $1
	`

	var model sharedRecapModel
	err := model.Scan(r.pool.QueryRow(ctx, query, recapID))
	if errors.Is(err, pgx.ErrNoRows) {
		return entity.SharedRecap{}, fmt.Errorf(
			"get shared recap for recap %s: %w",
			recapID,
			entity.ErrSharedRecapNotFound,
		)
	}
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("get shared recap by recap id: %w", err)
	}

	sharedRecap, err := sharedRecapModelToEntity(model)
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("convert shared recap model: %w", err)
	}

	return sharedRecap, nil
}

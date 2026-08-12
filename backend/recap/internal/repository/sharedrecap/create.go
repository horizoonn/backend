package sharedrecap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) Create(
	ctx context.Context,
	sharedRecap entity.SharedRecap,
) (entity.SharedRecapCreation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	model, err := sharedRecapToModel(sharedRecap)
	if err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("convert shared recap to model: %w", err)
	}

	const query = `
		INSERT INTO recap.shared_recaps (
			token,
			recap_id,
			snapshot
		)
		VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT unique_shared_recap_per_recap DO NOTHING
		RETURNING token, created_at
	`

	var creation entity.SharedRecapCreation
	err = r.pool.QueryRow(
		ctx,
		query,
		model.Token,
		model.RecapID,
		model.Snapshot,
	).Scan(&creation.Token, &creation.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		existing, getErr := r.getCreationByRecapID(ctx, model.RecapID)
		if getErr != nil {
			return entity.SharedRecapCreation{}, fmt.Errorf(
				"get existing shared recap after conflict: %w",
				getErr,
			)
		}

		return existing, nil
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "shared_recaps_recap_id_fkey" {
			return entity.SharedRecapCreation{}, fmt.Errorf(
				"create shared recap for recap %s: %w",
				model.RecapID,
				entity.ErrRecapNotFound,
			)
		}

		return entity.SharedRecapCreation{}, fmt.Errorf("create shared recap: %w", err)
	}

	creation.Created = true

	return creation, nil
}

func (r *Repository) getCreationByRecapID(
	ctx context.Context,
	recapID uuid.UUID,
) (entity.SharedRecapCreation, error) {
	const query = `
		SELECT token, created_at
		FROM recap.shared_recaps
		WHERE recap_id = $1
	`

	var creation entity.SharedRecapCreation
	if err := r.pool.QueryRow(ctx, query, recapID).Scan(
		&creation.Token,
		&creation.CreatedAt,
	); err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf(
			"get shared recap creation by recap id %s: %w",
			recapID,
			err,
		)
	}

	return creation, nil
}

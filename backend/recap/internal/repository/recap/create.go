package recap

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
	recap entity.Recap,
) (entity.RecapCreation, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	model, err := recapToModel(recap)
	if err != nil {
		return entity.RecapCreation{}, fmt.Errorf("convert recap to model: %w", err)
	}

	const query = `
		INSERT INTO recap.recaps (
			user_id,
			year,
			archetype,
			archetype_title,
			archetype_description,
			archetype_reasons,
			slides
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT ON CONSTRAINT unique_recap_per_user_year DO NOTHING
		RETURNING id
	`

	err = r.pool.QueryRow(
		ctx,
		query,
		model.UserID,
		model.Year,
		model.Archetype,
		model.ArchetypeTitle,
		model.ArchetypeDescription,
		model.ArchetypeReasons,
		model.Slides,
	).Scan(&model.ID)
	if errors.Is(err, pgx.ErrNoRows) {
		existingID, getErr := r.getIDByProfileAndYear(ctx, model.UserID, model.Year)
		if getErr != nil {
			return entity.RecapCreation{}, fmt.Errorf(
				"get existing recap after conflict: %w",
				getErr,
			)
		}

		return entity.RecapCreation{ID: existingID, Created: false}, nil
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23503" &&
			pgErr.ConstraintName == "recaps_user_id_fkey" {
			return entity.RecapCreation{}, fmt.Errorf(
				"create recap for profile %s: %w",
				model.UserID,
				entity.ErrProfileNotFound,
			)
		}

		return entity.RecapCreation{}, fmt.Errorf("create recap: %w", err)
	}

	return entity.RecapCreation{ID: model.ID, Created: true}, nil
}

func (r *Repository) getIDByProfileAndYear(
	ctx context.Context,
	profileID uuid.UUID,
	year int32,
) (uuid.UUID, error) {
	const query = `
		SELECT id
		FROM recap.recaps
		WHERE user_id = $1
		  AND year = $2
	`

	var id uuid.UUID
	if err := r.pool.QueryRow(ctx, query, profileID, year).Scan(&id); err != nil {
		return uuid.Nil, fmt.Errorf("get recap id: %w", err)
	}

	return id, nil
}

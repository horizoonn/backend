package profile

import (
	"context"
	"fmt"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) List(ctx context.Context) ([]entity.Profile, error) {
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
		ORDER BY created_at, id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list profiles: %w", err)
	}
	defer rows.Close()

	profiles := make([]entity.Profile, 0)
	for rows.Next() {
		var model profileModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan profile: %w", err)
		}

		profiles = append(profiles, profileModelToEntity(model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate profiles: %w", err)
	}

	return profiles, nil
}

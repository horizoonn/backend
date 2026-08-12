package listing

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) ListOldestActiveFavorites(
	ctx context.Context,
	userID uuid.UUID,
	period entity.Period,
	limit int,
) ([]entity.FavoriteListingPreview, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		SELECT
			l.id,
			l.title,
			l.price,
			l.category_id,
			f.created_at
		FROM recap.favorites AS f
		JOIN recap.listings AS l ON l.id = f.listing_id
		WHERE f.user_id = $1
		  AND f.created_at >= $2
		  AND f.created_at < $3
		  AND l.status = 'active'
		  AND l.seller_id <> $1
		  AND NOT EXISTS (
			SELECT 1
			FROM recap.deals AS d
			WHERE d.listing_id = l.id
			  AND d.completed_at IS NOT NULL
		  )
		ORDER BY f.created_at, l.id
		LIMIT $4
	`

	rows, err := r.pool.Query(ctx, query, userID, period.From, period.To, limit)
	if err != nil {
		return nil, fmt.Errorf("list oldest active favorites: %w", err)
	}
	defer rows.Close()

	result := make([]entity.FavoriteListingPreview, 0, limit)
	for rows.Next() {
		var model favoriteListingPreviewModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan oldest active favorite: %w", err)
		}

		result = append(result, favoriteListingPreviewModelToEntity(model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate oldest active favorites: %w", err)
	}

	return result, nil
}

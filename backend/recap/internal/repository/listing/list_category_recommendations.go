package listing

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) ListCategoryRecommendations(
	ctx context.Context,
	userID uuid.UUID,
	categoryID uuid.UUID,
	preferredSubcategoryID *uuid.UUID,
	limit int,
) ([]entity.ListingPreview, error) {
	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		SELECT
			l.id,
			l.title,
			l.price,
			l.category_id
		FROM recap.listings AS l
		WHERE l.category_id = $2
		  AND l.status = 'active'
		  AND l.seller_id <> $1
		  AND NOT EXISTS (
			SELECT 1
			FROM recap.favorites AS f
			WHERE f.user_id = $1
			  AND f.listing_id = l.id
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM recap.deals AS d
			WHERE d.listing_id = l.id
			  AND d.completed_at IS NOT NULL
		  )
		ORDER BY
			CASE
				WHEN $3::uuid IS NOT NULL AND l.subcategory_id = $3 THEN 0
				ELSE 1
			END,
			l.created_at DESC,
			l.id
		LIMIT $4
	`

	rows, err := r.pool.Query(
		ctx,
		query,
		userID,
		categoryID,
		preferredSubcategoryID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list category recommendations: %w", err)
	}
	defer rows.Close()

	result := make([]entity.ListingPreview, 0, limit)
	for rows.Next() {
		var model listingPreviewModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan category recommendation: %w", err)
		}

		result = append(result, listingPreviewModelToEntity(model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category recommendations: %w", err)
	}

	return result, nil
}

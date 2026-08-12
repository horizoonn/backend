package activity

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) ListActivityByCategories(
	ctx context.Context,
	userID uuid.UUID,
	period entity.Period,
) ([]entity.CategoryActivity, error) {
	if !period.Valid() {
		return nil, fmt.Errorf(
			"list activity by categories: invalid period [%v, %v)",
			period.From,
			period.To,
		)
	}

	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		WITH category_activity AS (
			SELECT
				l.category_id,
				l.subcategory_id,
				COUNT(*) AS views,
				0::bigint AS favorites,
				0::bigint AS purchases,
				0::bigint AS sales
			FROM recap.views AS v
			JOIN recap.listings AS l ON l.id = v.listing_id
			WHERE v.user_id = $1
			  AND v.viewed_at >= $2
			  AND v.viewed_at < $3
			GROUP BY l.category_id, l.subcategory_id

			UNION ALL

			SELECT
				l.category_id,
				l.subcategory_id,
				0::bigint AS views,
				COUNT(*) AS favorites,
				0::bigint AS purchases,
				0::bigint AS sales
			FROM recap.favorites AS f
			JOIN recap.listings AS l ON l.id = f.listing_id
			WHERE f.user_id = $1
			  AND f.created_at >= $2
			  AND f.created_at < $3
			GROUP BY l.category_id, l.subcategory_id

			UNION ALL

			SELECT
				l.category_id,
				l.subcategory_id,
				0::bigint AS views,
				0::bigint AS favorites,
				COUNT(*) AS purchases,
				0::bigint AS sales
			FROM recap.deals AS d
			JOIN recap.listings AS l ON l.id = d.listing_id
			WHERE d.buyer_id = $1
			  AND d.completed_at IS NOT NULL
			  AND d.completed_at >= $2
			  AND d.completed_at < $3
			GROUP BY l.category_id, l.subcategory_id

			UNION ALL

			SELECT
				l.category_id,
				l.subcategory_id,
				0::bigint AS views,
				0::bigint AS favorites,
				0::bigint AS purchases,
				COUNT(*) AS sales
			FROM recap.deals AS d
			JOIN recap.listings AS l ON l.id = d.listing_id
			WHERE l.seller_id = $1
			  AND d.completed_at IS NOT NULL
			  AND d.completed_at >= $2
			  AND d.completed_at < $3
			GROUP BY l.category_id, l.subcategory_id
		),
		category_totals AS (
			SELECT
				category_id,
				subcategory_id,
				SUM(views)::bigint AS views,
				SUM(favorites)::bigint AS favorites,
				SUM(purchases)::bigint AS purchases,
				SUM(sales)::bigint AS sales
			FROM category_activity
			GROUP BY category_id, subcategory_id
		)
		SELECT
			c.id AS category_id,
			c.title AS category_title,
			sc.id AS subcategory_id,
			sc.title AS subcategory_title,
			ct.views AS views,
			ct.favorites AS favorites,
			ct.purchases AS purchases,
			ct.sales AS sales
		FROM category_totals AS ct
		JOIN recap.categories AS c ON c.id = ct.category_id
		LEFT JOIN recap.subcategories AS sc
		       ON sc.id = ct.subcategory_id
		      AND sc.category_id = ct.category_id
		ORDER BY c.title, c.id, sc.title NULLS FIRST, sc.id NULLS FIRST
	`

	rows, err := r.pool.Query(ctx, query, userID, period.From, period.To)
	if err != nil {
		return nil, fmt.Errorf("list activity by categories: %w", err)
	}
	defer rows.Close()

	result := make([]entity.CategoryActivity, 0)
	for rows.Next() {
		var model categoryActivityModel
		if err := model.Scan(rows); err != nil {
			return nil, fmt.Errorf("scan category activity: %w", err)
		}

		result = append(result, categoryActivityModelToEntity(model))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate category activities: %w", err)
	}

	return result, nil
}

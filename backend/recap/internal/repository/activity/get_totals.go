package activity

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (r *Repository) GetActivityTotals(
	ctx context.Context,
	userID uuid.UUID,
	period entity.Period,
) (entity.UserActivity, error) {
	if !period.Valid() {
		return entity.UserActivity{}, fmt.Errorf(
			"get activity totals: invalid period [%v, %v)",
			period.From,
			period.To,
		)
	}

	ctx, cancel := context.WithTimeout(ctx, r.opTimeout)
	defer cancel()

	const query = `
		WITH activity_events AS (
			SELECT
				'view'::text AS event_type,
				v.viewed_at AS occurred_at,
				v.listing_id,
				l.category_id,
				0::bigint AS amount,
				FALSE AS listing_active
			FROM recap.views AS v
			JOIN recap.listings AS l ON l.id = v.listing_id
			WHERE v.user_id = $1
			  AND v.viewed_at >= $2
			  AND v.viewed_at < $3

			UNION ALL

			SELECT
				'favorite'::text AS event_type,
				f.created_at AS occurred_at,
				f.listing_id,
				l.category_id,
				0::bigint AS amount,
				l.status = 'active' AS listing_active
			FROM recap.favorites AS f
			JOIN recap.listings AS l ON l.id = f.listing_id
			WHERE f.user_id = $1
			  AND f.created_at >= $2
			  AND f.created_at < $3

			UNION ALL

			SELECT
				'purchase'::text AS event_type,
				d.completed_at AS occurred_at,
				d.listing_id,
				l.category_id,
				0::bigint AS amount,
				FALSE AS listing_active
			FROM recap.deals AS d
			JOIN recap.listings AS l ON l.id = d.listing_id
			WHERE d.buyer_id = $1
			  AND d.completed_at IS NOT NULL
			  AND d.completed_at >= $2
			  AND d.completed_at < $3

			UNION ALL

			SELECT
				'sale'::text AS event_type,
				d.completed_at AS occurred_at,
				d.listing_id,
				l.category_id,
				d.price AS amount,
				FALSE AS listing_active
			FROM recap.deals AS d
			JOIN recap.listings AS l ON l.id = d.listing_id
			WHERE l.seller_id = $1
			  AND d.completed_at IS NOT NULL
			  AND d.completed_at >= $2
			  AND d.completed_at < $3

			UNION ALL

			SELECT
				'message_as_buyer'::text AS event_type,
				m.created_at AS occurred_at,
				m.listing_id,
				NULL::uuid AS category_id,
				0::bigint AS amount,
				FALSE AS listing_active
			FROM recap.messages AS m
			WHERE m.buyer_id = $1
			  AND m.created_at >= $2
			  AND m.created_at < $3

			UNION ALL

			SELECT
				'message_as_seller'::text AS event_type,
				m.created_at AS occurred_at,
				m.listing_id,
				NULL::uuid AS category_id,
				0::bigint AS amount,
				FALSE AS listing_active
			FROM recap.messages AS m
			WHERE m.seller_id = $1
			  AND m.created_at >= $2
			  AND m.created_at < $3

			UNION ALL

			SELECT
				'listing_created'::text AS event_type,
				l.created_at AS occurred_at,
				l.id AS listing_id,
				l.category_id,
				0::bigint AS amount,
				FALSE AS listing_active
			FROM recap.listings AS l
			WHERE l.seller_id = $1
			  AND l.created_at >= $2
			  AND l.created_at < $3
		)
		SELECT
			COUNT(DISTINCT (occurred_at AT TIME ZONE 'UTC')::date) AS active_days,
			COUNT(*) FILTER (WHERE event_type = 'view') AS views,
			COUNT(DISTINCT listing_id) FILTER (
				WHERE event_type = 'view'
			) AS unique_listings_seen,
			COUNT(*) FILTER (WHERE event_type = 'favorite') AS favorites,
			COUNT(*) FILTER (
				WHERE event_type = 'favorite' AND listing_active
			) AS favorites_active,
			COUNT(*) FILTER (WHERE event_type = 'purchase') AS purchases,
			COUNT(*) FILTER (WHERE event_type = 'sale') AS sales,
			COALESCE(
				SUM(amount) FILTER (WHERE event_type = 'sale'),
				0
			)::bigint AS sales_amount,
			COUNT(*) FILTER (
				WHERE event_type = 'message_as_buyer'
			) AS messages_as_buyer,
			COUNT(*) FILTER (
				WHERE event_type = 'message_as_seller'
			) AS messages_as_seller,
			COUNT(DISTINCT category_id) FILTER (
				WHERE event_type IN ('view', 'favorite', 'purchase', 'sale')
			) AS categories_touched,
			COUNT(*) FILTER (
				WHERE event_type = 'listing_created'
			) AS listings_created
		FROM activity_events
	`

	var result entity.UserActivity
	err := r.pool.QueryRow(ctx, query, userID, period.From, period.To).Scan(
		&result.ActiveDays,
		&result.Views,
		&result.UniqueListingsSeen,
		&result.Favorites,
		&result.FavoritesActive,
		&result.Purchases,
		&result.Sales,
		&result.SalesAmount,
		&result.MessagesAsBuyer,
		&result.MessagesAsSeller,
		&result.CategoriesTouched,
		&result.ListingsCreated,
	)
	if err != nil {
		return entity.UserActivity{}, fmt.Errorf("get activity totals: %w", err)
	}

	return result, nil
}

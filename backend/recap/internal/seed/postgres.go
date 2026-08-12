package seed

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAlreadySeeded = errors.New("demo data already seeded")
	ErrPartialSeed   = errors.New("partial demo data found")
)

type Writer struct {
	pool *pgxpool.Pool
}

func NewWriter(pool *pgxpool.Pool) *Writer {
	return &Writer{pool: pool}
}

func (w *Writer) Write(ctx context.Context, dataset Dataset, reset bool) error {
	if len(dataset.Users) == 0 {
		return fmt.Errorf("write dataset: no users")
	}

	err := pgx.BeginFunc(ctx, w.pool, func(tx pgx.Tx) error {
		if err := prepareDemoUsers(ctx, tx, dataset.Users, reset); err != nil {
			return err
		}
		if err := upsertCatalog(ctx, tx, dataset); err != nil {
			return err
		}

		return copyDataset(ctx, tx, dataset)
	})
	if err != nil {
		return fmt.Errorf("write demo dataset: %w", err)
	}

	return nil
}

func prepareDemoUsers(ctx context.Context, tx pgx.Tx, users []UserRow, reset bool) error {
	userIDs := make([]uuid.UUID, len(users))
	for index, user := range users {
		userIDs[index] = user.ID
	}

	var existingUsers int
	if err := tx.QueryRow(
		ctx,
		`SELECT COUNT(*) FROM recap.users WHERE id = ANY($1)`,
		userIDs,
	).Scan(&existingUsers); err != nil {
		return fmt.Errorf("count existing demo users: %w", err)
	}

	switch {
	case reset:
		if _, err := tx.Exec(ctx, `DELETE FROM recap.users WHERE id = ANY($1)`, userIDs); err != nil {
			return fmt.Errorf("delete existing demo users: %w", err)
		}
	case existingUsers == len(users):
		return ErrAlreadySeeded
	case existingUsers > 0:
		return fmt.Errorf("%w: found %d of %d profiles", ErrPartialSeed, existingUsers, len(users))
	}

	return nil
}

func copyDataset(ctx context.Context, tx pgx.Tx, dataset Dataset) error {
	if err := copyUsers(ctx, tx, dataset.Users); err != nil {
		return err
	}
	if err := copyListings(ctx, tx, dataset.Listings); err != nil {
		return err
	}
	if err := copyViews(ctx, tx, dataset.Views); err != nil {
		return err
	}
	if err := copyFavorites(ctx, tx, dataset.Favorites); err != nil {
		return err
	}
	if err := copyMessages(ctx, tx, dataset.Messages); err != nil {
		return err
	}

	return copyDeals(ctx, tx, dataset.Deals)
}

func upsertCatalog(ctx context.Context, tx pgx.Tx, dataset Dataset) error {
	for _, category := range dataset.Categories {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO recap.categories (id, title)
			 VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET title = EXCLUDED.title`,
			category.ID,
			category.Title,
		); err != nil {
			return fmt.Errorf("upsert category %q: %w", category.Code, err)
		}
	}

	for _, subcategory := range dataset.Subcategories {
		if _, err := tx.Exec(
			ctx,
			`INSERT INTO recap.subcategories (id, category_id, title)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (id) DO UPDATE
			 SET category_id = EXCLUDED.category_id,
			     title = EXCLUDED.title`,
			subcategory.ID,
			subcategory.CategoryID,
			subcategory.Title,
		); err != nil {
			return fmt.Errorf("upsert subcategory %q: %w", subcategory.Code, err)
		}
	}

	return nil
}

func copyUsers(ctx context.Context, tx pgx.Tx, rows []UserRow) error {
	return copyRows(
		ctx,
		tx,
		"users",
		[]string{"id", "name", "surname", "avatar_url", "hint", "created_at"},
		len(rows),
		func(index int) ([]any, error) {
			row := rows[index]

			return []any{row.ID, row.Name, row.Surname, row.AvatarURL, row.Hint, row.RegisteredAt}, nil
		},
	)
}

func copyListings(ctx context.Context, tx pgx.Tx, rows []ListingRow) error {
	return copyRows(
		ctx,
		tx,
		"listings",
		[]string{
			"id",
			"seller_id",
			"title",
			"description",
			"price",
			"category_id",
			"subcategory_id",
			"status",
			"created_at",
			"closed_at",
		},
		len(rows),
		func(index int) ([]any, error) {
			row := rows[index]

			return []any{
				row.ID,
				row.SellerID,
				row.Title,
				row.Description,
				row.Price,
				row.CategoryID,
				row.SubcategoryID,
				row.Status,
				row.CreatedAt,
				row.ClosedAt,
			}, nil
		},
	)
}

func copyViews(ctx context.Context, tx pgx.Tx, rows []ViewRow) error {
	return copyRows(
		ctx,
		tx,
		"views",
		[]string{"id", "user_id", "listing_id", "viewed_at"},
		len(rows),
		func(index int) ([]any, error) {
			row := rows[index]

			return []any{row.ID, row.UserID, row.ListingID, row.ViewedAt}, nil
		},
	)
}

func copyFavorites(ctx context.Context, tx pgx.Tx, rows []FavoriteRow) error {
	return copyRows(
		ctx,
		tx,
		"favorites",
		[]string{"user_id", "listing_id", "created_at"},
		len(rows),
		func(index int) ([]any, error) {
			row := rows[index]

			return []any{row.UserID, row.ListingID, row.CreatedAt}, nil
		},
	)
}

func copyMessages(ctx context.Context, tx pgx.Tx, rows []MessageRow) error {
	return copyRows(
		ctx,
		tx,
		"messages",
		[]string{"id", "buyer_id", "seller_id", "listing_id", "created_at"},
		len(rows),
		func(index int) ([]any, error) {
			row := rows[index]

			return []any{row.ID, row.BuyerID, row.SellerID, row.ListingID, row.CreatedAt}, nil
		},
	)
}

func copyDeals(ctx context.Context, tx pgx.Tx, rows []DealRow) error {
	return copyRows(
		ctx,
		tx,
		"deals",
		[]string{"id", "listing_id", "buyer_id", "price", "created_at", "completed_at"},
		len(rows),
		func(index int) ([]any, error) {
			row := rows[index]

			return []any{row.ID, row.ListingID, row.BuyerID, row.Price, row.CreatedAt, row.CompletedAt}, nil
		},
	)
}

func copyRows(
	ctx context.Context,
	tx pgx.Tx,
	table string,
	columns []string,
	rowCount int,
	rowFn func(int) ([]any, error),
) error {
	if rowCount == 0 {
		return nil
	}

	inserted, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"recap", table},
		columns,
		pgx.CopyFromSlice(rowCount, rowFn),
	)
	if err != nil {
		return fmt.Errorf("copy %s: %w", table, err)
	}
	if inserted != int64(rowCount) {
		return fmt.Errorf("copy %s: inserted %d of %d rows", table, inserted, rowCount)
	}

	return nil
}

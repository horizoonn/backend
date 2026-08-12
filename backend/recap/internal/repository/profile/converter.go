package profile

import (
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func profileModelToEntity(model profileModel) entity.Profile {
	return entity.Profile{
		ID:           model.ID,
		Name:         model.Name,
		Surname:      model.Surname,
		AvatarURL:    nullableTextValue(model.AvatarURL),
		Hint:         nullableTextZeroValue(model.Hint),
		RegisteredAt: model.RegisteredAt,
	}
}

func nullableTextZeroValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func nullableTextValue(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

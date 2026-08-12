// Package sharedrecap implements public recap sharing scenarios.
package sharedrecap

import (
	"context"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

type (
	recapRepository interface {
		GetByID(ctx context.Context, id uuid.UUID) (entity.Recap, error)
	}

	profileRepository interface {
		GetByID(ctx context.Context, id uuid.UUID) (entity.Profile, error)
	}

	sharedRecapRepository interface {
		Create(
			ctx context.Context,
			sharedRecap entity.SharedRecap,
		) (entity.SharedRecapCreation, error)

		GetByToken(
			ctx context.Context,
			token entity.SharedRecapToken,
		) (entity.SharedRecap, error)

		GetByRecapID(
			ctx context.Context,
			recapID uuid.UUID,
		) (entity.SharedRecap, error)
	}

	TokenGenerator func() (entity.SharedRecapToken, error)
)

type sharedRecapService struct {
	recapRepository       recapRepository
	profileRepository     profileRepository
	sharedRecapRepository sharedRecapRepository
	tokenGenerator        TokenGenerator
}

func NewSharedRecapService(
	recapRepository recapRepository,
	profileRepository profileRepository,
	sharedRecapRepository sharedRecapRepository,
	tokenGenerator TokenGenerator,
) *sharedRecapService {
	return &sharedRecapService{
		recapRepository:       recapRepository,
		profileRepository:     profileRepository,
		sharedRecapRepository: sharedRecapRepository,
		tokenGenerator:        tokenGenerator,
	}
}

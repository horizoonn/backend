package sharedrecap

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (s *sharedRecapService) Share(
	ctx context.Context,
	recapID uuid.UUID,
) (entity.SharedRecapCreation, error) {
	if err := ctx.Err(); err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("share recap: %w", err)
	}
	if recapID == uuid.Nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("share recap: %w", entity.ErrRecapIDRequired)
	}

	existing, err := s.sharedRecapRepository.GetByRecapID(ctx, recapID)
	if err == nil {
		return entity.SharedRecapCreation{
			Token:     existing.Token,
			CreatedAt: existing.CreatedAt,
			Created:   false,
		}, nil
	}
	if !errors.Is(err, entity.ErrSharedRecapNotFound) {
		return entity.SharedRecapCreation{}, fmt.Errorf("get existing shared recap: %w", err)
	}

	recap, err := s.recapRepository.GetByID(ctx, recapID)
	if err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("get recap: %w", err)
	}

	profile, err := s.profileRepository.GetByID(ctx, recap.UserID)
	if err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("get recap profile: %w", err)
	}

	sharedRecap, err := buildSharedRecap(recap, profile)
	if err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("build shared recap: %w", err)
	}

	token, err := generateValidToken(s.tokenGenerator)
	if err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("generate shared recap token: %w", err)
	}
	sharedRecap.Token = token

	creation, err := s.sharedRecapRepository.Create(ctx, sharedRecap)
	if err != nil {
		return entity.SharedRecapCreation{}, fmt.Errorf("create shared recap: %w", err)
	}

	return creation, nil
}

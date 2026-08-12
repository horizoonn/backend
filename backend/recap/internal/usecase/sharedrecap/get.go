package sharedrecap

import (
	"context"
	"fmt"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (s *sharedRecapService) Get(
	ctx context.Context,
	token entity.SharedRecapToken,
) (entity.SharedRecap, error) {
	if err := ctx.Err(); err != nil {
		return entity.SharedRecap{}, fmt.Errorf("get shared recap: %w", err)
	}
	if err := validateToken(token); err != nil {
		return entity.SharedRecap{}, fmt.Errorf(
			"get shared recap: %w: %w",
			entity.ErrSharedRecapTokenInvalid,
			err,
		)
	}

	sharedRecap, err := s.sharedRecapRepository.GetByToken(ctx, token)
	if err != nil {
		return entity.SharedRecap{}, fmt.Errorf("get shared recap: %w", err)
	}

	return sharedRecap, nil
}

package recap

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (s *recapServer) OpenSimilarRecommendation(
	ctx context.Context,
	params recapapi.OpenSimilarRecommendationParams,
) (recapapi.OpenSimilarRecommendationRes, error) {
	destination, err := s.recapService.OpenSimilarRecommendation(
		ctx,
		uuid.UUID(params.ID),
		uuid.UUID(params.ListingId),
	)

	switch {
	case errors.Is(err, entity.ErrRecapIDRequired):
		apiErr := recapapi.OpenSimilarRecommendationBadRequest(
			s.errorBody(recapapi.ErrorCodeBadRequest, err.Error(), "open similar recommendation", err),
		)
		return &apiErr, nil
	case errors.Is(err, entity.ErrRecapNotFound),
		errors.Is(err, entity.ErrRecommendationNotFound):
		apiErr := recapapi.OpenSimilarRecommendationNotFound(
			s.errorBody(recapapi.ErrorCodeRecapNotFound, "recommendation not found", "open similar recommendation", err),
		)
		return &apiErr, nil
	case err != nil:
		apiErr := recapapi.OpenSimilarRecommendationInternalServerError(
			s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "open similar recommendation", err),
		)
		return &apiErr, nil
	}

	return &recapapi.Redirect{Location: destination}, nil
}

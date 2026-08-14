package recap

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (s *recapServer) OpenRecapAction(
	ctx context.Context,
	params recapapi.OpenRecapActionParams,
) (recapapi.OpenRecapActionRes, error) {
	destination, err := s.recapService.OpenAction(
		ctx,
		uuid.UUID(params.ID),
		string(params.Action),
	)

	switch {
	case errors.Is(err, entity.ErrRecapIDRequired),
		errors.Is(err, entity.ErrUnsupportedRecapAction):
		apiErr := recapapi.OpenRecapActionBadRequest(
			s.errorBody(recapapi.ErrorCodeBadRequest, err.Error(), "open recap action", err),
		)
		return &apiErr, nil
	case errors.Is(err, entity.ErrRecapNotFound),
		errors.Is(err, entity.ErrFavoriteCategoryNotFound):
		apiErr := recapapi.OpenRecapActionNotFound(
			s.errorBody(recapapi.ErrorCodeRecapNotFound, "recap action not found", "open recap action", err),
		)
		return &apiErr, nil
	case err != nil:
		apiErr := recapapi.OpenRecapActionInternalServerError(
			s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "open recap action", err),
		)
		return &apiErr, nil
	}

	return &recapapi.Redirect{Location: destination}, nil
}

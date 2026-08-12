package recap

import (
	"context"
	"errors"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/controller/http/converter"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (s *recapServer) GetSharedRecap(
	ctx context.Context,
	params recapapi.GetSharedRecapParams,
) (recapapi.GetSharedRecapRes, error) {
	sharedRecap, err := s.sharedRecapService.Get(
		ctx,
		entity.SharedRecapToken(params.Token),
	)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrSharedRecapTokenInvalid):
			apiErr := recapapi.GetSharedRecapBadRequest(
				s.errorBody(recapapi.ErrorCodeBadRequest, err.Error(), "get shared recap", err),
			)
			return &apiErr, nil

		case errors.Is(err, entity.ErrSharedRecapNotFound):
			apiErr := recapapi.GetSharedRecapNotFound(
				s.errorBody(
					recapapi.ErrorCodeSharedRecapNotFound,
					err.Error(),
					"get shared recap",
					err,
				),
			)
			return &apiErr, nil

		default:
			apiErr := recapapi.GetSharedRecapInternalServerError(
				s.errorBody(
					recapapi.ErrorCodeInternalError,
					"internal error",
					"get shared recap",
					err,
				),
			)
			return &apiErr, nil
		}
	}

	response, err := converter.ConvertEntitySharedRecapToAPISharedRecap(sharedRecap)
	if err != nil {
		apiErr := recapapi.GetSharedRecapInternalServerError(
			s.errorBody(
				recapapi.ErrorCodeInternalError,
				"internal error",
				"encode shared recap",
				err,
			),
		)

		return &apiErr, nil
	}

	return &response, nil
}

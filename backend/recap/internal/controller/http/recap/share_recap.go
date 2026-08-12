package recap

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func (s *recapServer) ShareRecap(
	ctx context.Context,
	params recapapi.ShareRecapParams,
) (recapapi.ShareRecapRes, error) {
	creation, err := s.sharedRecapService.Share(ctx, uuid.UUID(params.ID))

	switch {
	case errors.Is(err, entity.ErrRecapIDRequired):
		apiErr := recapapi.ShareRecapBadRequest(
			s.errorBody(recapapi.ErrorCodeBadRequest, err.Error(), "share recap", err),
		)

		return &apiErr, nil

	case errors.Is(err, entity.ErrRecapNotFound),
		errors.Is(err, entity.ErrProfileNotFound):
		apiErr := recapapi.ShareRecapNotFound(
			s.errorBody(
				recapapi.ErrorCodeRecapNotFound,
				entity.ErrRecapNotFound.Error(),
				"share recap",
				err,
			),
		)

		return &apiErr, nil

	case err != nil:
		apiErr := recapapi.ShareRecapInternalServerError(
			s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "share recap", err),
		)

		return &apiErr, nil
	}

	publicURL := s.publicBaseURL.JoinPath(
		"shared-recaps",
		string(creation.Token),
	)
	response := recapapi.SharedRecapLink{
		Token:     recapapi.SharedRecapToken(creation.Token),
		URL:       *publicURL,
		CreatedAt: creation.CreatedAt,
	}

	if creation.Created {
		created := recapapi.ShareRecapCreated(response)

		return &created, nil
	}

	existing := recapapi.ShareRecapOK(response)

	return &existing, nil
}

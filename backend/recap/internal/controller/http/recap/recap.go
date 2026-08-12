// Package recap exposes recap delivery handlers.
package recap

import (
	"context"
	"errors"
	"net/url"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/controller/http/converter"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

type (
	recapService interface {
		Create(ctx context.Context, profileID uuid.UUID, year int) (entity.RecapCreation, error)
		Get(ctx context.Context, recapID uuid.UUID) (entity.Recap, error)
	}

	profileService interface {
		List(ctx context.Context) ([]entity.Profile, error)
	}

	sharedRecapService interface {
		Share(ctx context.Context, recapID uuid.UUID) (entity.SharedRecapCreation, error)
		Get(ctx context.Context, token entity.SharedRecapToken) (entity.SharedRecap, error)
	}
)

type recapServer struct {
	recapapi.UnimplementedHandler

	logger             *zap.Logger
	recapService       recapService
	profileService     profileService
	sharedRecapService sharedRecapService
	publicBaseURL      url.URL
}

var _ recapapi.Handler = (*recapServer)(nil)

func NewRecapServer(
	logger *zap.Logger,
	recapService recapService,
	profileService profileService,
	sharedRecapService sharedRecapService,
	publicBaseURL url.URL,
) *recapServer {
	return &recapServer{
		logger:             logger,
		recapService:       recapService,
		profileService:     profileService,
		sharedRecapService: sharedRecapService,
		publicBaseURL:      publicBaseURL,
	}
}

func (s *recapServer) ListProfiles(ctx context.Context) (recapapi.ListProfilesRes, error) {
	profiles, err := s.profileService.List(ctx)
	if err != nil {
		apiErr := s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "list profiles", err)

		return &apiErr, nil
	}

	items := make([]recapapi.Profile, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, converter.ConvertEntityProfileToAPIProfile(profile))
	}

	return &recapapi.ListProfilesOK{Items: items}, nil
}

func (s *recapServer) CreateRecap(
	ctx context.Context,
	req *recapapi.CreateRecapRequest,
) (recapapi.CreateRecapRes, error) {
	creation, err := s.recapService.Create(ctx, uuid.UUID(req.ProfileId), int(req.Year))

	switch {
	case errors.Is(err, entity.ErrProfileIDRequired):
		apiErr := recapapi.CreateRecapBadRequest(
			s.errorBody(recapapi.ErrorCodeBadRequest, err.Error(), "create recap", err),
		)

		return &apiErr, nil

	case errors.Is(err, entity.ErrProfileNotFound):
		apiErr := recapapi.CreateRecapNotFound(
			s.errorBody(recapapi.ErrorCodeProfileNotFound, err.Error(), "create recap", err),
		)

		return &apiErr, nil

	case errors.Is(err, entity.ErrNotEnoughActivity):
		apiErr := recapapi.CreateRecapConflict(
			s.errorBody(recapapi.ErrorCodeNotEnoughActivity, err.Error(), "create recap", err),
		)

		return &apiErr, nil

	case err != nil:
		apiErr := recapapi.CreateRecapInternalServerError(
			s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "create recap", err),
		)

		return &apiErr, nil
	}

	response := recapapi.CreateRecapResponse{ID: recapapi.UUID(creation.ID)}

	if creation.Created {
		created := recapapi.CreateRecapCreated(response)

		return &created, nil
	}

	existing := recapapi.CreateRecapOK(response)

	return &existing, nil
}

func (s *recapServer) GetRecap(
	ctx context.Context,
	params recapapi.GetRecapParams,
) (recapapi.GetRecapRes, error) {
	recap, err := s.recapService.Get(ctx, uuid.UUID(params.ID))

	switch {
	case errors.Is(err, entity.ErrRecapIDRequired):
		apiErr := recapapi.GetRecapBadRequest(
			s.errorBody(recapapi.ErrorCodeBadRequest, err.Error(), "get recap", err),
		)

		return &apiErr, nil

	case errors.Is(err, entity.ErrRecapNotFound):
		apiErr := recapapi.GetRecapNotFound(
			s.errorBody(recapapi.ErrorCodeRecapNotFound, err.Error(), "get recap", err),
		)

		return &apiErr, nil

	case err != nil:
		apiErr := recapapi.GetRecapInternalServerError(
			s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "get recap", err),
		)

		return &apiErr, nil
	}

	response, err := converter.ConvertEntityRecapToAPIRecap(recap)
	if err != nil {
		apiErr := recapapi.GetRecapInternalServerError(
			s.errorBody(recapapi.ErrorCodeInternalError, "internal error", "encode recap", err),
		)

		return &apiErr, nil
	}

	return &response, nil
}

// errorBody builds the single error shape of the contract and writes the real
// cause to the log, so a client report with a requestId is enough to find it.
func (s *recapServer) errorBody(
	code recapapi.ErrorCode,
	message string,
	operation string,
	cause error,
) recapapi.Error {
	requestID := uuid.New()
	logger := s.logger.WithOptions(zap.AddCallerSkip(1))
	fields := []zap.Field{
		zap.Error(cause),
		zap.String("code", string(code)),
		zap.String("requestId", requestID.String()),
	}

	if code == recapapi.ErrorCodeInternalError {
		logger.Error(operation, fields...)
	} else {
		logger.Warn(operation, fields...)
	}

	return recapapi.Error{
		Code:      code,
		Message:   message,
		RequestId: recapapi.NewOptUUID(recapapi.UUID(requestID)),
	}
}

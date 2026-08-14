package app

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/config"
	recapcontroller "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/controller/http/recap"
	activityrepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/activity"
	listingrepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/listing"
	profilerepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/profile"
	recaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/recap"
	sharedrecaprepo "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/repository/sharedrecap"
	profileusecase "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/profile"
	recapusecase "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/recap"
	sharedrecapusecase "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/sharedrecap"
)

const apiPrefix = "/api/v1"

func buildHTTPHandler(
	pool *pgxpool.Pool,
	cfg config.Config,
	logger *zap.Logger,
) (http.Handler, error) {
	publicBaseURL, err := cfg.Public.URL()
	if err != nil {
		return nil, fmt.Errorf("parse public base url: %w", err)
	}

	activityRepository := activityrepo.New(
		pool,
		cfg.Repository.OperationTimeout,
	)

	listingRepository := listingrepo.New(
		pool,
		cfg.Repository.OperationTimeout,
	)

	profileRepository := profilerepo.New(
		pool,
		cfg.Repository.OperationTimeout,
	)

	recapRepository := recaprepo.New(
		pool,
		cfg.Repository.OperationTimeout,
	)

	sharedRecapRepository := sharedrecaprepo.New(
		pool,
		cfg.Repository.OperationTimeout,
	)

	recapService := recapusecase.NewRecapService(
		activityRepository,
		recapRepository,
		profileRepository,
		listingRepository,
	)

	profileService := profileusecase.NewProfileService(
		profileRepository,
	)

	sharedRecapService := sharedrecapusecase.NewSharedRecapService(
		recapRepository,
		profileRepository,
		sharedRecapRepository,
		sharedrecapusecase.GenerateToken,
	)

	recapHandler := recapcontroller.NewRecapServer(
		logger,
		recapService,
		profileService,
		sharedRecapService,
		publicBaseURL,
	)

	server, err := recapapi.NewServer(recapHandler)
	if err != nil {
		return nil, fmt.Errorf("create recap openapi server: %w", err)
	}

	mux := http.NewServeMux()
	registerHealthEndpoints(mux, pool, cfg.Repository.OperationTimeout)
	mux.Handle(apiPrefix+"/", http.StripPrefix(apiPrefix, server))

	return mux, nil
}

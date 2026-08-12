package sharedrecap

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	sharedrecapmocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/sharedrecap/mocks"
)

func TestNewSharedRecapService(t *testing.T) {
	t.Parallel()

	recapRepository := sharedrecapmocks.NewMockRecapRepository(t)
	profileRepository := sharedrecapmocks.NewMockProfileRepository(t)
	sharedRecapRepository := sharedrecapmocks.NewMockSharedRecapRepository(t)
	tokenGenerator := func() (entity.SharedRecapToken, error) {
		return validSharedRecapToken(), nil
	}

	service := NewSharedRecapService(
		recapRepository,
		profileRepository,
		sharedRecapRepository,
		tokenGenerator,
	)

	require.Same(t, recapRepository, service.recapRepository)
	require.Same(t, profileRepository, service.profileRepository)
	require.Same(t, sharedRecapRepository, service.sharedRecapRepository)
	require.NotNil(t, service.tokenGenerator)
}

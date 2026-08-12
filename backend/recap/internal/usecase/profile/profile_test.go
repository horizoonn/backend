package profile

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	profilemocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/profile/mocks"
)

func TestNewProfileService(t *testing.T) {
	t.Parallel()

	repository := profilemocks.NewMockProfileRepository(t)

	service := NewProfileService(repository)

	require.Same(t, repository, service.profileRepository)
}

func TestProfileService_List(t *testing.T) {
	t.Parallel()

	storageError := errors.New("profile storage unavailable")
	profiles := []entity.Profile{
		{ID: uuid.New(), Name: "Анна", Surname: "Смирнова"},
		{ID: uuid.New(), Name: "Иван", Surname: "Петров"},
	}

	type args struct {
		cancelContext bool
	}

	tests := []struct {
		name            string
		args            args
		want            []entity.Profile
		wantErr         error
		wantErrContains string
		setupMocks      func(context.Context, *profilemocks.MockProfileRepository)
	}{
		{
			name:    "canceled context",
			args:    args{cancelContext: true},
			wantErr: context.Canceled,
		},
		{
			name: "repository error",
			setupMocks: func(ctx context.Context, repository *profilemocks.MockProfileRepository) {
				repository.EXPECT().
					List(ctx).
					Return(nil, storageError).
					Once()
			},
			wantErr:         storageError,
			wantErrContains: "list profiles",
		},
		{
			name: "empty profile list",
			setupMocks: func(ctx context.Context, repository *profilemocks.MockProfileRepository) {
				repository.EXPECT().
					List(ctx).
					Return([]entity.Profile{}, nil).
					Once()
			},
			want: []entity.Profile{},
		},
		{
			name: "success",
			setupMocks: func(ctx context.Context, repository *profilemocks.MockProfileRepository) {
				repository.EXPECT().
					List(ctx).
					Return(profiles, nil).
					Once()
			},
			want: profiles,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := t.Context()
			if tt.args.cancelContext {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			repository := profilemocks.NewMockProfileRepository(t)
			if tt.setupMocks != nil {
				tt.setupMocks(ctx, repository)
			}
			service := NewProfileService(repository)

			got, err := service.List(ctx)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

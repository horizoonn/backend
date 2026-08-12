package sharedrecap

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
	sharedrecapmocks "github.com/avito-hackaton-team/avito-recap/backend/recap/internal/usecase/sharedrecap/mocks"
)

func TestSharedRecapService_Get(t *testing.T) {
	t.Parallel()

	token := validSharedRecapToken()
	stored := validSharedRecap(uuid.New())
	storageError := errors.New("shared recap storage unavailable")

	type args struct {
		token         entity.SharedRecapToken
		cancelContext bool
	}

	tests := []struct {
		name            string
		args            args
		want            entity.SharedRecap
		wantErr         error
		wantErrContains string
		setupMocks      func(context.Context, *sharedrecapmocks.MockSharedRecapRepository)
	}{
		{
			name:    "canceled context",
			args:    args{token: token, cancelContext: true},
			wantErr: context.Canceled,
		},
		{
			name:            "token has invalid length",
			args:            args{token: "short"},
			wantErr:         entity.ErrSharedRecapTokenInvalid,
			wantErrContains: "token must contain 22 characters",
		},
		{
			name: "token contains invalid character",
			args: args{token: entity.SharedRecapToken(
				strings.Repeat("a", entity.SharedRecapTokenLength-1) + "/",
			)},
			wantErr:         entity.ErrSharedRecapTokenInvalid,
			wantErrContains: "invalid character",
		},
		{
			name: "repository error",
			args: args{token: token},
			setupMocks: func(ctx context.Context, repository *sharedrecapmocks.MockSharedRecapRepository) {
				repository.EXPECT().
					GetByToken(ctx, token).
					Return(entity.SharedRecap{}, storageError).
					Once()
			},
			wantErr:         storageError,
			wantErrContains: "get shared recap",
		},
		{
			name: "success",
			args: args{token: token},
			setupMocks: func(ctx context.Context, repository *sharedrecapmocks.MockSharedRecapRepository) {
				repository.EXPECT().
					GetByToken(ctx, token).
					Return(stored, nil).
					Once()
			},
			want: stored,
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

			service, dependencies := newSharedRecapTestService(t, GenerateToken)
			if tt.setupMocks != nil {
				tt.setupMocks(ctx, dependencies.sharedRecap)
			}

			got, err := service.Get(ctx, tt.args.token)

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

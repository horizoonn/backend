package recap

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestRecapService_Get(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	stored := entity.Recap{ID: recapID, UserID: uuid.New(), Year: maxRecapYear}
	storageError := errors.New("recap lookup failed")

	type args struct {
		recapID       uuid.UUID
		cancelContext bool
	}

	tests := []struct {
		name       string
		args       args
		want       entity.Recap
		wantErr    error
		setupMocks func(context.Context, recapTestDependencies)
	}{
		{
			name:    "canceled context",
			args:    args{recapID: recapID, cancelContext: true},
			wantErr: context.Canceled,
		},
		{
			name:    "missing recap id",
			args:    args{recapID: uuid.Nil},
			wantErr: entity.ErrRecapIDRequired,
		},
		{
			name: "repository error",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				dependencies.recap.EXPECT().
					GetByID(ctx, recapID).
					Return(entity.Recap{}, storageError).
					Once()
			},
			wantErr: storageError,
		},
		{
			name: "success",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies recapTestDependencies) {
				dependencies.recap.EXPECT().
					GetByID(ctx, recapID).
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

			service, dependencies := newRecapTestService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(ctx, dependencies)
			}

			got, err := service.Get(ctx, tt.args.recapID)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

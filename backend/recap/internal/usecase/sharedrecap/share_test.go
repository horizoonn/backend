package sharedrecap

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestSharedRecapService_Share(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	profileID := uuid.New()
	recap := validRecapForSharing(recapID, profileID)
	profile := entity.Profile{ID: profileID, Name: "Анна", Surname: "Смирнова"}
	token := validSharedRecapToken()
	createdAt := time.Date(2026, time.December, 20, 12, 0, 0, 0, time.UTC)
	creation := entity.SharedRecapCreation{Token: token, CreatedAt: createdAt, Created: true}
	storageError := errors.New("shared recap storage unavailable")
	recapError := errors.New("recap storage unavailable")
	profileError := errors.New("profile storage unavailable")
	tokenError := errors.New("random source unavailable")

	type args struct {
		recapID       uuid.UUID
		cancelContext bool
	}

	tests := []struct {
		name            string
		args            args
		tokenGenerator  TokenGenerator
		want            entity.SharedRecapCreation
		wantErr         error
		wantErrContains string
		setupMocks      func(context.Context, sharedRecapTestDependencies)
	}{
		{
			name:    "canceled context",
			args:    args{recapID: recapID, cancelContext: true},
			wantErr: context.Canceled,
		},
		{
			name:            "recap id is required",
			args:            args{recapID: uuid.Nil},
			wantErr:         entity.ErrRecapIDRequired,
			wantErrContains: "share recap",
		},
		{
			name: "existing public recap is returned idempotently",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				dependencies.sharedRecap.EXPECT().
					GetByRecapID(ctx, recapID).
					Return(entity.SharedRecap{Token: token, CreatedAt: createdAt}, nil).
					Once()
			},
			want: entity.SharedRecapCreation{Token: token, CreatedAt: createdAt, Created: false},
		},
		{
			name: "existing public recap lookup fails",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				dependencies.sharedRecap.EXPECT().
					GetByRecapID(ctx, recapID).
					Return(entity.SharedRecap{}, storageError).
					Once()
			},
			wantErr:         storageError,
			wantErrContains: "get existing shared recap",
		},
		{
			name: "recap lookup fails",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				dependencies.sharedRecap.EXPECT().
					GetByRecapID(ctx, recapID).
					Return(entity.SharedRecap{}, entity.ErrSharedRecapNotFound).
					Once()
				dependencies.recap.EXPECT().
					GetByID(ctx, recapID).
					Return(entity.Recap{}, recapError).
					Once()
			},
			wantErr:         recapError,
			wantErrContains: "get recap",
		},
		{
			name: "profile lookup fails",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				dependencies.sharedRecap.EXPECT().
					GetByRecapID(ctx, recapID).
					Return(entity.SharedRecap{}, entity.ErrSharedRecapNotFound).
					Once()
				dependencies.recap.EXPECT().
					GetByID(ctx, recapID).
					Return(recap, nil).
					Once()
				dependencies.profile.EXPECT().
					GetByID(ctx, profileID).
					Return(entity.Profile{}, profileError).
					Once()
			},
			wantErr:         profileError,
			wantErrContains: "get recap profile",
		},
		{
			name: "snapshot cannot be built",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				invalidRecap := recap
				invalidRecap.Slides = []byte(`[]`)
				expectShareBuildInputs(ctx, dependencies, invalidRecap, profile)
			},
			wantErrContains: "build shared recap: extract recap facts: active days slide is missing",
		},
		{
			name: "token generator fails",
			args: args{recapID: recapID},
			tokenGenerator: func() (entity.SharedRecapToken, error) {
				return "", tokenError
			},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				expectShareBuildInputs(ctx, dependencies, recap, profile)
			},
			wantErr:         tokenError,
			wantErrContains: "generate shared recap token",
		},
		{
			name: "token generator returns invalid token",
			args: args{recapID: recapID},
			tokenGenerator: func() (entity.SharedRecapToken, error) {
				return "short", nil
			},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				expectShareBuildInputs(ctx, dependencies, recap, profile)
			},
			wantErrContains: "token must contain 22 characters",
		},
		{
			name: "create public recap fails",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				expectShareBuildInputs(ctx, dependencies, recap, profile)
				dependencies.sharedRecap.EXPECT().
					Create(ctx, mock.Anything).
					Return(entity.SharedRecapCreation{}, storageError).
					Once()
			},
			wantErr:         storageError,
			wantErrContains: "create shared recap",
		},
		{
			name: "success",
			args: args{recapID: recapID},
			setupMocks: func(ctx context.Context, dependencies sharedRecapTestDependencies) {
				expectShareBuildInputs(ctx, dependencies, recap, profile)
				dependencies.sharedRecap.EXPECT().
					Create(ctx, mock.MatchedBy(func(snapshot entity.SharedRecap) bool {
						return snapshot.Token == token && snapshot.RecapID == recapID
					})).
					Return(creation, nil).
					Once()
			},
			want: creation,
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
			tokenGenerator := tt.tokenGenerator
			if tokenGenerator == nil {
				tokenGenerator = func() (entity.SharedRecapToken, error) { return token, nil }
			}
			service, dependencies := newSharedRecapTestService(t, tokenGenerator)
			if tt.setupMocks != nil {
				tt.setupMocks(ctx, dependencies)
			}

			got, err := service.Share(ctx, tt.args.recapID)

			switch {
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			case tt.wantErrContains == "":
				require.NoError(t, err)
			default:
				require.Error(t, err)
			}
			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestSharedRecapService_Share_PersistsPublicSnapshot(t *testing.T) {
	t.Parallel()

	recapID := uuid.New()
	profileID := uuid.New()
	recap := validRecapForSharing(recapID, profileID)
	profile := entity.Profile{
		ID:      profileID,
		Name:    "Анна",
		Surname: "Секретная фамилия",
		Hint:    "Приватная подсказка",
	}
	token := validSharedRecapToken()
	creation := entity.SharedRecapCreation{Token: token, Created: true}
	service, dependencies := newSharedRecapTestService(t, func() (entity.SharedRecapToken, error) {
		return token, nil
	})
	ctx := t.Context()
	expectShareBuildInputs(ctx, dependencies, recap, profile)

	wantSnapshot := validSharedRecap(recapID)
	dependencies.sharedRecap.EXPECT().
		Create(ctx, wantSnapshot).
		Return(creation, nil).
		Once()

	got, err := service.Share(ctx, recapID)

	require.NoError(t, err)
	require.Equal(t, creation, got)
}

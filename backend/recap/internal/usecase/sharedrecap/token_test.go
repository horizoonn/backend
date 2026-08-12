package sharedrecap

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/internal/entity"
)

func TestGenerateToken(t *testing.T) {
	t.Parallel()

	token, err := GenerateToken()
	require.NoError(t, err)
	require.Len(t, token, entity.SharedRecapTokenLength)
	require.NoError(t, validateToken(token))
}

func TestGenerateValidToken(t *testing.T) {
	t.Parallel()

	token := validSharedRecapToken()
	generatorError := errors.New("random source unavailable")
	tests := []struct {
		name            string
		generator       TokenGenerator
		want            entity.SharedRecapToken
		wantErr         error
		wantErrContains string
	}{
		{name: "success", generator: func() (entity.SharedRecapToken, error) { return token, nil }, want: token},
		{
			name: "generator error",
			generator: func() (entity.SharedRecapToken, error) {
				return "", generatorError
			},
			wantErr: generatorError,
		},
		{
			name: "invalid generated token",
			generator: func() (entity.SharedRecapToken, error) {
				return "short", nil
			},
			wantErrContains: "token must contain 22 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := generateValidToken(tt.generator)

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

func TestValidateToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		token           entity.SharedRecapToken
		wantErrContains string
	}{
		{
			name:  "lowercase",
			token: entity.SharedRecapToken(strings.Repeat("a", 22)),
		},
		{
			name:  "all allowed character classes",
			token: "abcXYZ019_-abcXYZ019_-",
		},
		{
			name:            "too short",
			token:           "short",
			wantErrContains: "token must contain 22 characters",
		},
		{
			name:            "too long",
			token:           entity.SharedRecapToken(strings.Repeat("a", 23)),
			wantErrContains: "token must contain 22 characters",
		},
		{
			name:            "padding is forbidden",
			token:           "abcXYZ019_-abcXYZ019_=",
			wantErrContains: `invalid character '='`,
		},
		{
			name:            "slash is forbidden",
			token:           "abcXYZ019_-abcXYZ019_/",
			wantErrContains: `invalid character '/'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateToken(tt.token)

			if tt.wantErrContains != "" {
				require.ErrorContains(t, err, tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIsURLSafeBase64Character(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		char rune
		want bool
	}{
		{name: "lowercase", char: 'a', want: true},
		{name: "uppercase", char: 'Z', want: true},
		{name: "digit", char: '7', want: true},
		{name: "underscore", char: '_', want: true},
		{name: "hyphen", char: '-', want: true},
		{name: "padding", char: '='},
		{name: "slash", char: '/'},
		{name: "unicode", char: 'я'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, isURLSafeBase64Character(tt.char))
		})
	}
}

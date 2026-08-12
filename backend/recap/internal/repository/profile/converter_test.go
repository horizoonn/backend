package profile

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

func TestNullableTextValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value pgtype.Text
		want  *string
	}{
		{
			name:  "null",
			value: pgtype.Text{},
		},
		{
			name:  "empty string",
			value: pgtype.Text{String: "", Valid: true},
			want:  stringPointer(""),
		},
		{
			name:  "non-empty string",
			value: pgtype.Text{String: "profile hint", Valid: true},
			want:  stringPointer("profile hint"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, nullableTextValue(tt.value))
		})
	}
}

func TestNullableTextZeroValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value pgtype.Text
		want  string
	}{
		{
			name:  "null",
			value: pgtype.Text{},
		},
		{
			name:  "empty string",
			value: pgtype.Text{String: "", Valid: true},
		},
		{
			name:  "non-empty string",
			value: pgtype.Text{String: "profile hint", Valid: true},
			want:  "profile hint",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, nullableTextZeroValue(tt.value))
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

package connection

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProjectsValidateKeysNormalisesAndDeduplicates(t *testing.T) {
	got, err := ValidateProjectKeys([]any{" proj ", "DEMO", "proj", "A_1"})
	require.NoError(t, err)
	require.Equal(t, []string{"PROJ", "DEMO", "A_1"}, got)
}

func TestProjectsValidateKeysAcceptsAnEmptyList(t *testing.T) {
	got, err := ValidateProjectKeys([]any{})
	require.NoError(t, err)
	require.Empty(t, got)
}

func TestProjectsValidateKeysRejectsBadInput(t *testing.T) {
	tooMany := make([]any, 101)
	for i := range tooMany {
		tooMany[i] = fmt.Sprintf("P%d", i)
	}
	for name, raw := range map[string]any{
		"missing array":     nil,
		"not an array":      "PROJ",
		"a non-string":      []any{"PROJ", 7},
		"a lower digit key": []any{"1PROJ"},
		"a dash":            []any{"PRO-J"},
		"too long":          []any{"A23456789012345678901234567"},
		"empty key":         []any{" "},
		"more than 100":     tooMany,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateProjectKeys(raw)
			require.ErrorIs(t, err, ErrInvalidProjects)
			out := Classify(err)
			require.Equal(t, CodeValidation, out.Code)
			require.Equal(t, FieldProjectKeys, out.Field)
		})
	}
}

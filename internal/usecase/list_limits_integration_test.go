package usecase_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/usecase"
)

// Public search is reachable without an account, so it refuses inputs that
// would make the query expensive.
func TestPublicListRejectsAbusiveInput(t *testing.T) {
	e := setup(t)
	guest := usecase.Actor{}

	cases := map[string]usecase.ListPropertiesInput{
		"long search":     {Search: strings.Repeat("a", 101), Page: 1, Limit: 20},
		"deep page":       {Page: 502, Limit: 20},
		"many facilities": {FacilityIDs: make([]uuid.UUID, 21), Page: 1, Limit: 20},
	}
	for label, in := range cases {
		if _, _, err := e.prop.List(ctx(), guest, in); statusOf(err) != 422 {
			t.Errorf("%s: %d, want 422", label, statusOf(err))
		}
	}
	// Limits at the boundary are accepted.
	if _, _, err := e.prop.List(ctx(), guest, usecase.ListPropertiesInput{Search: strings.Repeat("a", 100), Page: 501, Limit: 20}); err != nil {
		t.Errorf("boundary input refused: %v", err)
	}
}

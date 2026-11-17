package usecase_test

import (
	"strings"
	"testing"

	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/validator"
)

func TestListingExtrasValidation(t *testing.T) {
	v := validator.New()
	eleven := make([]string, 11)
	for i := range eleven {
		eleven[i] = "Area"
	}
	tooFar := []entity.NearbyPlace{{Name: "Pasar", Meters: 1_000_001}}
	blank := []entity.NearbyPlace{{Name: "", Meters: 10}}
	long := []string{strings.Repeat("a", 51)}
	noFurnishing, badFurnishing := "", "mewah"

	cases := map[string]struct {
		in      any
		wantErr bool
	}{
		"empty create":           {usecase.PropertyDetails{}, false},
		"valid extras":           {usecase.PropertyDetails{ExtraOutdoor: []string{"Rooftop"}, ExtraDistances: []entity.NearbyPlace{{Name: "Pasar", Meters: 800}}}, false},
		"too many outdoor":       {usecase.PropertyDetails{ExtraOutdoor: eleven}, true},
		"outdoor name too long":  {usecase.PropertyDetails{ExtraOutdoor: long}, true},
		"distance out of range":  {usecase.PropertyDetails{ExtraDistances: tooFar}, true},
		"distance without name":  {usecase.PropertyDetails{ExtraDistances: blank}, true},
		"patch leaves unchanged": {usecase.PropertyDetailsPatch{}, false},
		"patch clears":           {usecase.PropertyDetailsPatch{ExtraOutdoor: &[]string{}, ExtraDistances: &[]entity.NearbyPlace{}}, false},
		"patch too many outdoor": {usecase.PropertyDetailsPatch{ExtraOutdoor: &eleven}, true},
		"patch distance too far": {usecase.PropertyDetailsPatch{ExtraDistances: &tooFar}, true},
		"patch empty furnishing": {usecase.PropertyDetailsPatch{FurnishingStatus: &noFurnishing}, false},
		"patch bad furnishing":   {usecase.PropertyDetailsPatch{FurnishingStatus: &badFurnishing}, true},
	}
	for name, c := range cases {
		if got := v.Validate(c.in) != nil; got != c.wantErr {
			t.Errorf("%s: error=%v, want %v (%v)", name, got, c.wantErr, v.Validate(c.in))
		}
	}
}

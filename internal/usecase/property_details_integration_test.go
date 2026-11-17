package usecase_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/usecase"
)

func TestListingDetailsAreStoredAndPublic(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	price := 1_400_000_000.0
	fee := 1_000_000.0
	school := 100

	in := usecase.CreatePropertyInput{
		Owner: usecase.OwnerInput{Name: "Owner", Phone: ownerPhone},
		Code:  "D-" + uuid.NewString()[:12], Name: "Rumah Detail", Type: "house", TransactionType: "sale", Price: &price,
		GeneralLocation: "Cibubur", StreetName: streetName,
	}
	in.ShortDescription = "Dijual rumah yang asri dan siap huni"
	in.SubDistrict, in.PostalCode, in.Orientation = "Ciangsana", "32345", "Timur"
	in.MaintenanceFee = &fee
	in.CertSHM = true
	in.MasterBedrooms, in.ChildBedrooms, in.MasterBathrooms = 3, 2, 3
	in.ElectricityCapacity, in.AirConditioners, in.WaterSource = "2200 VA", 1, "PAM"
	in.FurnishingStatus = entity.FurnishingFurnished
	in.HasPool, in.HasGarage = true, true
	in.DistSchool = &school

	p, err := e.prop.Create(ctx(), a, in, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}

	// A guest sees every detail field but no owner or restricted address data.
	v, err := e.prop.GetByID(ctx(), usecase.Actor{}, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	pub, ok := v.(usecase.PublicProperty)
	if !ok {
		t.Fatalf("guest view is %T, want PublicProperty", v)
	}
	if !pub.CertSHM || pub.CertHGB || pub.MasterBedrooms != 3 || pub.ChildBedrooms != 2 || pub.PostalCode != "32345" ||
		pub.FurnishingStatus != "furnished" || !pub.HasPool || pub.DistSchool == nil || *pub.DistSchool != 100 ||
		pub.MaintenanceFee == nil || *pub.MaintenanceFee != fee || pub.Orientation != "Timur" {
		t.Errorf("details missing from public view: %+v", pub.PropertyDetailsView)
	}
	if body := mustJSON(t, pub); strings.Contains(body, ownerPhone) || strings.Contains(body, streetName) {
		t.Error("public view leaks owner phone or restricted address")
	}

	// Partial update only touches the supplied fields.
	pool, kitchens := false, 2
	upd, err := e.prop.Update(ctx(), a, p.ID, usecase.UpdatePropertyInput{
		PropertyDetailsPatch: usecase.PropertyDetailsPatch{HasPool: &pool, Kitchens: &kitchens},
	}, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if upd.HasPool || upd.Kitchens != 2 || !upd.HasGarage || upd.MasterBedrooms != 3 {
		t.Errorf("partial update wrong: %+v", upd.PropertyDetailsView)
	}
}

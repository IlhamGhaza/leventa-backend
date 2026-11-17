package usecase_test

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/storage"
	"github.com/leventa-backend/internal/usecase"
)

func TestListingExtrasAreCleanedStoredAndPublic(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	price := 1_000_000_000.0
	in := usecase.CreatePropertyInput{
		Owner: usecase.OwnerInput{Name: "Owner", Phone: ownerPhone},
		Code:  "X-" + uuid.NewString()[:12], Name: "Rumah Extra", Type: "house", TransactionType: "sale", Price: &price,
		GeneralLocation: "Cibubur", StreetName: streetName,
	}
	in.ExtraOutdoor = []string{" Rooftop ", "rooftop", "", "Gazebo"}
	in.ExtraDistances = []entity.NearbyPlace{{Name: "Pasar", Meters: 800}, {Name: "pasar", Meters: 5}}
	p, err := e.prop.Create(ctx(), a, in, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}

	v, err := e.prop.GetByID(ctx(), usecase.Actor{}, p.ID)
	if err != nil {
		t.Fatal(err)
	}
	pub := v.(usecase.PublicProperty)
	if len(pub.ExtraOutdoor) != 2 || pub.ExtraOutdoor[0] != "Rooftop" || pub.ExtraOutdoor[1] != "Gazebo" {
		t.Errorf("extra_outdoor = %q, want [Rooftop Gazebo]", pub.ExtraOutdoor)
	}
	if len(pub.ExtraDistances) != 1 || pub.ExtraDistances[0] != (entity.NearbyPlace{Name: "Pasar", Meters: 800}) {
		t.Errorf("extra_distances = %+v, want [{Pasar 800}]", pub.ExtraDistances)
	}

	// nil leaves the lists alone; an empty list clears them.
	kitchens := 1
	upd, err := e.prop.Update(ctx(), a, p.ID, usecase.UpdatePropertyInput{
		PropertyDetailsPatch: usecase.PropertyDetailsPatch{Kitchens: &kitchens},
	}, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if len(upd.ExtraOutdoor) != 2 || len(upd.ExtraDistances) != 1 {
		t.Errorf("unrelated patch changed extras: %+v", upd.PropertyDetailsView)
	}
	upd, err = e.prop.Update(ctx(), a, p.ID, usecase.UpdatePropertyInput{
		PropertyDetailsPatch: usecase.PropertyDetailsPatch{ExtraOutdoor: &[]string{}, ExtraDistances: &[]entity.NearbyPlace{}},
	}, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if body := mustJSON(t, upd); !strings.Contains(body, `"extra_outdoor":[]`) || !strings.Contains(body, `"extra_distances":[]`) {
		t.Errorf("cleared extras should render [], got %s", body)
	}

	// A listing created without extras also renders [] (never null).
	plain := e.listing(t, a)
	v, err = e.prop.GetByID(ctx(), usecase.Actor{}, plain)
	if err != nil {
		t.Fatal(err)
	}
	if body := mustJSON(t, v); !strings.Contains(body, `"extra_outdoor":[]`) {
		t.Errorf("listing without extras: %s", body)
	}
}

func pngBytes(t *testing.T) []byte {
	t.Helper()
	var b bytes.Buffer
	if err := png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestThumbnailSelection(t *testing.T) {
	e := setup(t)
	uc := usecase.NewPropertyUsecase(pgRepo.NewPropertyRepository(e.db), pgRepo.NewPropertyImageRepository(e.db),
		pgRepo.NewOwnerRepository(e.db), pgRepo.NewCoListingRepository(e.db), pgRepo.NewMasterDataRepository(e.db),
		pgRepo.NewPropertyDocumentRepository(e.db), pgRepo.NewTransactor(e.db),
		storage.NewLocalStorage(t.TempDir(), 1<<20), e.audit())
	primary := e.user(t, entity.RoleSales, "active")
	co := e.user(t, entity.RoleSales, "active")
	other := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, primary)
	if _, err := e.co.Add(ctx(), admin, id, co.ID, usecase.RequestMeta{}); err != nil {
		t.Fatal(err)
	}
	up := func(pid uuid.UUID) *entity.PropertyImage {
		f, h := upload(t, "foto.png", pngBytes(t))
		img, err := uc.UploadImage(ctx(), primary, pid, f, h)
		if err != nil {
			t.Fatal(err)
		}
		return img
	}
	cover := func() (string, []uuid.UUID) {
		var p entity.Property
		if err := e.db.First(&p, "id = ?", id).Error; err != nil {
			t.Fatal(err)
		}
		var ids []uuid.UUID
		e.db.Model(&entity.PropertyImage{}).Where("property_id = ? AND is_thumbnail", id).Pluck("id", &ids)
		return p.Thumbnail, ids
	}

	first, second := up(id), up(id)
	if url, ids := cover(); url != first.URL || len(ids) != 1 || ids[0] != first.ID {
		t.Fatalf("first upload should be the cover: url=%q ids=%v", url, ids)
	}

	// Only the primary marketing may pick the cover.
	for name, actor := range map[string]usecase.Actor{"co-listing": co, "other sales": other} {
		if err := uc.SetThumbnail(ctx(), actor, id, second.ID); statusOf(err) != 403 && statusOf(err) != 404 {
			t.Errorf("%s set thumbnail: %d, want 403/404", name, statusOf(err))
		}
	}
	// An image of another listing is refused.
	foreign := up(e.listing(t, primary))
	if err := uc.SetThumbnail(ctx(), primary, id, foreign.ID); statusOf(err) != 404 {
		t.Errorf("foreign image: %d, want 404", statusOf(err))
	}

	if err := uc.SetThumbnail(ctx(), primary, id, second.ID); err != nil {
		t.Fatal(err)
	}
	if url, ids := cover(); url != second.URL || len(ids) != 1 || ids[0] != second.ID {
		t.Fatalf("cover not switched: url=%q ids=%v", url, ids)
	}

	// Deleting the cover promotes the remaining photo; deleting the last clears it.
	if err := uc.DeleteImage(ctx(), primary, id, second.ID); err != nil {
		t.Fatal(err)
	}
	if url, ids := cover(); url != first.URL || len(ids) != 1 || ids[0] != first.ID {
		t.Fatalf("cover not promoted: url=%q ids=%v", url, ids)
	}
	if err := uc.DeleteImage(ctx(), primary, id, first.ID); err != nil {
		t.Fatal(err)
	}
	if url, ids := cover(); url != "" || len(ids) != 0 {
		t.Fatalf("cover not cleared: url=%q ids=%v", url, ids)
	}
}

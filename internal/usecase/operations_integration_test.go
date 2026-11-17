package usecase_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/internal/usecase"
)

func (e *env) audit() *service.AuditService {
	return service.NewAuditService(pgRepo.NewAuditLogRepository(e.db))
}

type failingCommissionRepo struct {
	repository.CommissionRepository
}

func (failingCommissionRepo) Create(_ context.Context, _ *entity.Commission) error {
	return errors.New("boom")
}

func (e *env) dealUC(commissionRepo repository.CommissionRepository) *usecase.DealUsecase {
	if commissionRepo == nil {
		commissionRepo = pgRepo.NewCommissionRepository(e.db)
	}
	return usecase.NewDealUsecase(pgRepo.NewPropertyRepository(e.db), pgRepo.NewDealRepository(e.db), commissionRepo,
		pgRepo.NewTransactor(e.db), e.audit(), 2.5)
}

func statusOfProperty(e *env, id uuid.UUID) string {
	var p entity.Property
	e.db.Select("status").First(&p, "id = ?", id)
	return p.Status
}

func TestRecordDealAdminOnlyAndCommission(t *testing.T) {
	e := setup(t)
	owner := e.user(t, entity.RoleSales, "active")
	other := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, owner)
	uc := e.dealUC(nil)
	m := usecase.RequestMeta{}
	in := usecase.RecordDealInput{Type: "sale", Price: 1_000_000_000, BuyerName: "Pak Pembeli"}

	// Only Admin may record; marketing (even the primary owner) and guests may not.
	if _, err := uc.RecordDeal(ctx(), owner, id, in, m); statusOf(err) != 403 {
		t.Errorf("primary marketing: %d, want 403", statusOf(err))
	}
	if _, err := uc.RecordDeal(ctx(), other, id, in, m); statusOf(err) != 403 {
		t.Errorf("other marketing: %d, want 403", statusOf(err))
	}
	if _, err := uc.RecordDeal(ctx(), usecase.Actor{}, id, in, m); statusOf(err) != 401 {
		t.Errorf("guest: %d, want 401", statusOf(err))
	}
	if _, err := uc.Cancel(ctx(), owner, id, usecase.CancelInput{}, m); statusOf(err) != 403 {
		t.Errorf("marketing cancel: %d, want 403", statusOf(err))
	}
	// Type must match the listing.
	bad := in
	bad.Type = "rent"
	if _, err := uc.RecordDeal(ctx(), admin, id, bad, m); statusOf(err) != 422 {
		t.Errorf("type mismatch: %d, want 422", statusOf(err))
	}

	res, err := uc.RecordDeal(ctx(), admin, id, in, m)
	if err != nil {
		t.Fatal(err)
	}
	if got := statusOfProperty(e, id); got != "sold" {
		t.Errorf("status = %s, want sold", got)
	}
	if res.Deal.SalesID != owner.ID || res.Deal.RecordedBy != admin.ID {
		t.Errorf("deal owner/recorder wrong: %+v", res.Deal)
	}
	c := res.Commission
	if c.SalesID != owner.ID || c.Percentage != 2.5 || c.Amount != 25_000_000 || c.Status != "pending" {
		t.Fatalf("commission = %+v, want 2.5%% / 25,000,000 for the primary marketing", c)
	}
	// A second outcome is refused.
	if _, err := uc.RecordDeal(ctx(), admin, id, in, m); statusOf(err) != 409 {
		t.Errorf("second deal: %d, want 409", statusOf(err))
	}
	if _, err := uc.Cancel(ctx(), admin, id, usecase.CancelInput{}, m); statusOf(err) != 409 {
		t.Errorf("cancel after deal: %d, want 409", statusOf(err))
	}

	// Deal visibility: own listing only for marketing.
	if _, err := uc.Get(ctx(), other, res.Deal.ID); statusOf(err) != 404 {
		t.Errorf("other marketing deal: %d, want 404", statusOf(err))
	}
	if _, err := uc.Get(ctx(), owner, res.Deal.ID); err != nil {
		t.Errorf("own deal: %v", err)
	}
}

func TestRentDealSetsRented(t *testing.T) {
	e := setup(t)
	owner := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	rent := 10_000_000.0
	p, err := e.prop.Create(ctx(), owner, usecase.CreatePropertyInput{
		Owner: usecase.OwnerInput{Name: "Own"}, Code: "R-" + uuid.NewString()[:12], Name: "Ruko Sewa",
		Type: "shop_house", TransactionType: "rent", RentalPrice: &rent,
	}, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := e.dealUC(nil).RecordDeal(ctx(), admin, p.ID, usecase.RecordDealInput{Type: "rent", Price: 120_000_000, BuyerName: "Penyewa"}, usecase.RequestMeta{}); err != nil {
		t.Fatal(err)
	}
	if got := statusOfProperty(e, p.ID); got != "rented" {
		t.Errorf("status = %s, want rented", got)
	}
}

func TestCancelArchivesAndHidesFromLists(t *testing.T) {
	e := setup(t)
	owner := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, owner)
	uc := e.dealUC(nil)

	d, err := uc.Cancel(ctx(), admin, id, usecase.CancelInput{Notes: "owner batal jual"}, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if d.Outcome != "cancelled" || d.Price != nil || d.BuyerName != nil {
		t.Errorf("cancelled deal carries deal fields: %+v", d)
	}
	if got := statusOfProperty(e, id); got != "archived" {
		t.Fatalf("status = %s, want archived", got)
	}
	var commissions int64
	e.db.Model(&entity.Commission{}).Where("deal_id = ?", d.ID).Count(&commissions)
	if commissions != 0 {
		t.Errorf("cancellation created %d commissions", commissions)
	}

	// Hidden from the public list, the marketing list and the admin list.
	for label, actor := range map[string]usecase.Actor{"guest": {}, "owner": owner, "admin": admin} {
		in := usecase.ListPropertiesInput{Search: "Rumah Test", Page: 1, Limit: 100}
		if label == "owner" {
			in.Scope = "mine"
		}
		items, _, err := e.prop.List(ctx(), actor, in)
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range items {
			if it.ID == id {
				t.Errorf("%s list still contains the cancelled listing", label)
			}
		}
	}
	// Guests get 404 on the detail.
	if _, err := e.prop.GetByID(ctx(), usecase.Actor{}, id); statusOf(err) != 404 {
		t.Errorf("guest detail of cancelled listing: %d, want 404", statusOf(err))
	}
}

func TestDealRollsBackWhenCommissionFails(t *testing.T) {
	e := setup(t)
	owner := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, owner)

	_, err := e.dealUC(failingCommissionRepo{}).RecordDeal(ctx(), admin, id,
		usecase.RecordDealInput{Type: "sale", Price: 1_000_000_000, BuyerName: "Pembeli"}, usecase.RequestMeta{})
	if statusOf(err) != 500 {
		t.Fatalf("err status = %d, want 500", statusOf(err))
	}
	if got := statusOfProperty(e, id); got != "available" {
		t.Errorf("property status = %s after failure, want available", got)
	}
	var deals int64
	e.db.Model(&entity.Deal{}).Where("property_id = ?", id).Count(&deals)
	if deals != 0 {
		t.Errorf("deal row survived rollback: %d", deals)
	}
}

func TestConcurrentOutcomesRecordOnlyOne(t *testing.T) {
	e := setup(t)
	owner := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, owner)
	uc := e.dealUC(nil)

	var wg sync.WaitGroup
	results := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				_, err := uc.RecordDeal(ctx(), admin, id, usecase.RecordDealInput{Type: "sale", Price: 1_000_000, BuyerName: "Pembeli"}, usecase.RequestMeta{})
				results <- err
				return
			}
			_, err := uc.Cancel(ctx(), admin, id, usecase.CancelInput{}, usecase.RequestMeta{})
			results <- err
		}(i)
	}
	wg.Wait()
	close(results)
	ok := 0
	for err := range results {
		if err == nil {
			ok++
		}
	}
	if ok != 1 {
		t.Errorf("%d outcomes succeeded, want exactly 1", ok)
	}
	var deals int64
	e.db.Model(&entity.Deal{}).Where("property_id = ?", id).Count(&deals)
	if deals != 1 {
		t.Errorf("deal rows = %d, want 1", deals)
	}
}

func TestCommissionPolicyAndScope(t *testing.T) {
	e := setup(t)
	owner := e.user(t, entity.RoleSales, "active")
	other := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	res, err := e.dealUC(nil).RecordDeal(ctx(), admin, e.listing(t, owner),
		usecase.RecordDealInput{Type: "sale", Price: 1_000_000_000, BuyerName: "Pembeli"}, usecase.RequestMeta{})
	if err != nil {
		t.Fatal(err)
	}
	c := res.Commission

	uc := usecase.NewCommissionUsecase(pgRepo.NewCommissionRepository(e.db), e.audit())
	m := usecase.RequestMeta{}
	if _, err := uc.Get(ctx(), other, c.ID); statusOf(err) != 404 {
		t.Errorf("other marketing: %d, want 404", statusOf(err))
	}
	if _, err := uc.Get(ctx(), usecase.Actor{}, c.ID); statusOf(err) != 403 {
		t.Errorf("guest: %d, want 403", statusOf(err))
	}
	if _, err := uc.Get(ctx(), owner, c.ID); err != nil {
		t.Errorf("own commission: %v", err)
	}
	// Only admin may change percentage / status.
	if _, err := uc.UpdatePercentage(ctx(), owner, c.ID, 50, m); statusOf(err) != 403 {
		t.Errorf("marketing override: %d, want 403", statusOf(err))
	}
	if _, err := uc.UpdateStatus(ctx(), owner, c.ID, "paid", m); statusOf(err) != 403 {
		t.Errorf("marketing status: %d, want 403", statusOf(err))
	}
	upd, err := uc.UpdatePercentage(ctx(), admin, c.ID, 3, m)
	if err != nil || upd.Amount != 30_000_000 {
		t.Fatalf("admin override: %v %+v", err, upd)
	}
	if _, err := uc.UpdatePercentage(ctx(), admin, c.ID, 101, m); statusOf(err) != 422 {
		t.Errorf("pct 101: %d", statusOf(err))
	}
	if _, err := uc.UpdateStatus(ctx(), admin, c.ID, "paid", m); err != nil {
		t.Fatal(err)
	}
	if _, err := uc.UpdateStatus(ctx(), admin, c.ID, "cancelled", m); statusOf(err) != 409 {
		t.Errorf("paid -> cancelled: %d, want 409", statusOf(err))
	}
	if _, err := uc.UpdatePercentage(ctx(), admin, c.ID, 4, m); statusOf(err) != 409 {
		t.Errorf("override after paid: %d, want 409", statusOf(err))
	}
}

func ptr[T any](v T) *T { return &v }

func TestLocationAndFacilities(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	cust := usecase.Actor{} // guest: no account

	prov := &entity.Province{Name: "Prov " + a.ID.String()[:8]}
	prov2 := &entity.Province{Name: "Prov2 " + a.ID.String()[:8]}
	e.db.Create(prov)
	e.db.Create(prov2)
	city := &entity.City{ProvinceID: prov.ID, Name: "City"}
	e.db.Create(city)
	dist := &entity.District{CityID: city.ID, Name: "Dist"}
	e.db.Create(dist)
	otherCity := &entity.City{ProvinceID: prov2.ID, Name: "Other"}
	e.db.Create(otherCity)
	fac := &entity.Facility{Name: "Fac " + a.ID.String()[:8]}
	e.db.Create(fac)

	price := 1.0
	base := func() usecase.CreatePropertyInput {
		return usecase.CreatePropertyInput{
			Owner: usecase.OwnerInput{Name: "Own"}, Code: "L-" + a.ID.String()[:8] + time.Now().Format("150405.000"),
			Name: "Rumah Lokasi", Type: "house", TransactionType: "sale", Price: &price,
		}
	}
	m := usecase.RequestMeta{}

	bad := base()
	bad.ProvinceID, bad.CityID = &prov2.ID, &city.ID
	if _, err := e.prop.Create(ctx(), a, bad, m); statusOf(err) != 422 {
		t.Errorf("city/province mismatch: %d, want 422", statusOf(err))
	}
	bad = base()
	bad.CityID, bad.DistrictID = &otherCity.ID, &dist.ID
	if _, err := e.prop.Create(ctx(), a, bad, m); statusOf(err) != 422 {
		t.Errorf("district/city mismatch: %d, want 422", statusOf(err))
	}
	bad = base()
	bad.FacilityIDs = []uuid.UUID{uuid.New()}
	if _, err := e.prop.Create(ctx(), a, bad, m); statusOf(err) != 422 {
		t.Errorf("unknown facility: %d, want 422", statusOf(err))
	}

	ok := base()
	ok.ProvinceID, ok.CityID, ok.DistrictID, ok.FacilityIDs = &prov.ID, &city.ID, &dist.ID, []uuid.UUID{fac.ID}
	p, err := e.prop.Create(ctx(), a, ok, m)
	if err != nil {
		t.Fatal(err)
	}
	if p.District == nil || len(p.Facilities) != 1 {
		t.Fatalf("district/facilities missing: %+v", p.PublicProperty)
	}

	items, _, err := e.prop.List(ctx(), cust, usecase.ListPropertiesInput{DistrictID: &dist.ID, Page: 1, Limit: 10})
	if err != nil || len(items) != 1 || items[0].ID != p.ID {
		t.Errorf("district filter: %v %d", err, len(items))
	}

	// Replace facilities with none.
	empty := []uuid.UUID{}
	upd, err := e.prop.Update(ctx(), a, p.ID, usecase.UpdatePropertyInput{FacilityIDs: &empty}, m)
	if err != nil || len(upd.Facilities) != 0 {
		t.Errorf("clear facilities: %v %d", err, len(upd.Facilities))
	}
}

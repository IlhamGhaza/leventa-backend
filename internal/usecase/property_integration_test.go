package usecase_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/errs"
	gormPg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// Integration tests need a migrated PostgreSQL database:
//
//	TEST_DATABASE_URL=postgres://... go test ./internal/usecase/...
//
// They only create uniquely named rows and never truncate, but use a scratch database.

type env struct {
	db   *gorm.DB
	prop *usecase.PropertyUsecase
	own  *usecase.OwnerUsecase
	co   *usecase.CoListingUsecase
}

func setup(t *testing.T) *env {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	db, err := gorm.Open(gormPg.Open(url), &gorm.Config{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(20)
	t.Cleanup(func() { _ = sqlDB.Close() })

	propertyRepo := pgRepo.NewPropertyRepository(db)
	imageRepo := pgRepo.NewPropertyImageRepository(db)
	ownerRepo := pgRepo.NewOwnerRepository(db)
	coRepo := pgRepo.NewCoListingRepository(db)
	userRepo := pgRepo.NewUserRepository(db)
	tx := pgRepo.NewTransactor(db)
	audit := service.NewAuditService(pgRepo.NewAuditLogRepository(db))

	return &env{
		db:   db,
		prop: usecase.NewPropertyUsecase(propertyRepo, imageRepo, ownerRepo, coRepo, pgRepo.NewMasterDataRepository(db), pgRepo.NewPropertyDocumentRepository(db), tx, nil, audit),
		own:  usecase.NewOwnerUsecase(propertyRepo, coRepo, ownerRepo, audit),
		co:   usecase.NewCoListingUsecase(propertyRepo, coRepo, userRepo, tx, audit),
	}
}

func (e *env) user(t *testing.T, role, status string) usecase.Actor {
	t.Helper()
	u := &entity.User{
		Name: "T " + role, Email: uuid.NewString() + "@test.local", Role: role, Status: status,
		PasswordHash: "x", WhatsAppNumber: "6281200000099",
	}
	if err := e.db.Create(u).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return usecase.Actor{ID: u.ID, Role: role}
}

const (
	ownerPhone = "0811-SECRET-PHONE"
	streetName = "Jl. Rahasia Pribadi"
)

func (e *env) listing(t *testing.T, owner usecase.Actor) uuid.UUID {
	t.Helper()
	price := 1_500_000_000.0
	p, err := e.prop.Create(context.Background(), owner, usecase.CreatePropertyInput{
		Owner:           usecase.OwnerInput{Name: "Owner Rahasia", Email: "secret@owner.test", Phone: ownerPhone, PrivateAddress: "Alamat Pemilik Rahasia"},
		Code:            "T-" + uuid.NewString()[:12],
		Name:            "Rumah Test",
		Type:            "house",
		TransactionType: "sale",
		Price:           &price,
		GeneralLocation: "Kebayoran Baru",
		StreetName:      streetName,
		HouseNumber:     "77",
	}, usecase.RequestMeta{})
	if err != nil {
		t.Fatalf("create listing: %v", err)
	}
	return p.ID
}

func ctx() context.Context { return context.Background() }

func statusOf(err error) int {
	if err == nil {
		return 200
	}
	if ae, ok := errs.IsAppError(err); ok {
		return ae.Code
	}
	return 500
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestCreateSetsPrimaryFromActor(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	id := e.listing(t, a)

	var p entity.Property
	if err := e.db.First(&p, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if p.PrimaryMarketingID != a.ID {
		t.Fatalf("primary = %s, want %s", p.PrimaryMarketingID, a.ID)
	}
	var n int64
	e.db.Model(&entity.PropertyMarketingAssignment{}).Where("property_id = ? AND marketing_id = ? AND assignment_type = 'primary'", id, a.ID).Count(&n)
	if n != 1 {
		t.Fatalf("primary assignment rows = %d", n)
	}
}

func TestOnlySalesCanCreate(t *testing.T) {
	e := setup(t)
	for _, role := range []string{entity.RoleAdmin} {
		a := e.user(t, role, "active")
		price := 1.0
		_, err := e.prop.Create(ctx(), a, usecase.CreatePropertyInput{
			Owner: usecase.OwnerInput{Name: "Own"}, Code: "X-" + uuid.NewString()[:8], Name: "Nm", Type: "house",
			TransactionType: "sale", Price: &price,
		}, usecase.RequestMeta{})
		if statusOf(err) != 403 {
			t.Errorf("role %s: status %d, want 403", role, statusOf(err))
		}
	}
}

func TestEditAuthorization(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	b := e.user(t, entity.RoleSales, "active")
	co := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, a)
	if _, err := e.co.Add(ctx(), admin, id, co.ID, usecase.RequestMeta{}); err != nil {
		t.Fatal(err)
	}

	name := "Hacked"
	for label, actor := range map[string]usecase.Actor{"other marketing": b, "co-listing": co, "admin": admin} {
		_, err := e.prop.Update(ctx(), actor, id, usecase.UpdatePropertyInput{Name: &name}, usecase.RequestMeta{})
		if statusOf(err) != 403 {
			t.Errorf("%s update: status %d, want 403", label, statusOf(err))
		}
		if _, err := e.prop.Delete(ctx(), actor, id, usecase.RequestMeta{}); statusOf(err) != 403 {
			t.Errorf("%s delete: status %d, want 403", label, statusOf(err))
		}
		if err := e.prop.UpdateStatus(ctx(), actor, id, "archived", usecase.RequestMeta{}); statusOf(err) != 403 {
			t.Errorf("%s status: status %d, want 403", label, statusOf(err))
		}
	}
	if _, err := e.prop.Update(ctx(), usecase.Actor{}, id, usecase.UpdatePropertyInput{Name: &name}, usecase.RequestMeta{}); statusOf(err) != 401 {
		t.Errorf("anonymous update: %d, want 401", statusOf(err))
	}

	// primary can edit, and the name really changed only for the primary.
	good := "Nama Baru"
	if _, err := e.prop.Update(ctx(), a, id, usecase.UpdatePropertyInput{Name: &good}, usecase.RequestMeta{}); err != nil {
		t.Fatalf("primary update: %v", err)
	}
	var p entity.Property
	e.db.First(&p, "id = ?", id)
	if p.Name != good || p.PrimaryMarketingID != a.ID {
		t.Fatalf("unexpected state: %+v", p)
	}
}

func TestOwnerPrivacy(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	b := e.user(t, entity.RoleSales, "active")
	co := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, a)
	if _, err := e.co.Add(ctx(), admin, id, co.ID, usecase.RequestMeta{}); err != nil {
		t.Fatal(err)
	}

	// Primary reads owner.
	o, err := e.own.GetByProperty(ctx(), a, id)
	if err != nil || o.Phone != ownerPhone {
		t.Fatalf("primary owner read: %v %+v", err, o)
	}
	// Everyone else is denied, directly.
	for label, actor := range map[string]usecase.Actor{"other marketing": b, "co-listing": co, "admin": admin} {
		if _, err := e.own.GetByProperty(ctx(), actor, id); statusOf(err) != 403 {
			t.Errorf("%s owner read: %d, want 403", label, statusOf(err))
		}
		phone := "1"
		if _, err := e.own.UpdateByProperty(ctx(), actor, id, usecase.UpdateOwnerInput{Phone: &phone}, usecase.RequestMeta{}); statusOf(err) != 403 {
			t.Errorf("%s owner update: %d, want 403", label, statusOf(err))
		}
	}
	// Primary of ANOTHER listing cannot read this owner (IDOR).
	otherID := e.listing(t, b)
	if _, err := e.own.GetByProperty(ctx(), a, otherID); statusOf(err) != 403 {
		t.Errorf("cross-listing owner read: %d, want 403", statusOf(err))
	}
}

func TestResponsePayloadsNeverLeakPII(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	b := e.user(t, entity.RoleSales, "active")
	co := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	cust := usecase.Actor{} // guest: no account
	id := e.listing(t, a)
	if _, err := e.co.Add(ctx(), admin, id, co.ID, usecase.RequestMeta{}); err != nil {
		t.Fatal(err)
	}

	secrets := []string{ownerPhone, "Owner Rahasia", "secret@owner.test", "Alamat Pemilik Rahasia", streetName, `"house_number"`, `"latitude"`, `"address"`}

	for label, actor := range map[string]usecase.Actor{
		"anonymous": {}, "guest": cust, "other marketing": b, "admin": admin,
	} {
		got, err := e.prop.GetByID(ctx(), actor, id)
		if err != nil {
			t.Fatalf("%s detail: %v", label, err)
		}
		body := mustJSON(t, got)
		for _, s := range secrets {
			if strings.Contains(body, s) {
				t.Errorf("%s detail leaks %q: %s", label, s, body)
			}
		}
	}

	// Co-listing has the primary's view (exact address) but never owner PII.
	cod, err := e.prop.GetByID(ctx(), co, id)
	if err != nil {
		t.Fatal(err)
	}
	cob := mustJSON(t, cod)
	if !strings.Contains(cob, streetName) || !strings.Contains(cob, `"access_level":"co_listing"`) {
		t.Errorf("co-listing should see the exact address: %s", cob)
	}
	for _, sec := range []string{ownerPhone, "Owner Rahasia", "secret@owner.test", "Alamat Pemilik Rahasia"} {
		if strings.Contains(cob, sec) {
			t.Errorf("co-listing view leaks owner PII %q", sec)
		}
	}

	// Primary sees restricted address but still no owner PII.
	got, err := e.prop.GetByID(ctx(), a, id)
	if err != nil {
		t.Fatal(err)
	}
	body := mustJSON(t, got)
	if !strings.Contains(body, streetName) {
		t.Errorf("primary should see street name: %s", body)
	}
	for _, s := range []string{ownerPhone, "Owner Rahasia", "secret@owner.test", "Alamat Pemilik Rahasia"} {
		if strings.Contains(body, s) {
			t.Errorf("primary property detail leaks owner PII %q", s)
		}
	}

	// Lists never leak either, for any role.
	for label, actor := range map[string]usecase.Actor{"anonymous": {}, "guest": cust, "admin": admin} {
		items, _, err := e.prop.List(ctx(), actor, usecase.ListPropertiesInput{Search: "Rumah Test", Page: 1, Limit: 100})
		if err != nil {
			t.Fatal(err)
		}
		body := mustJSON(t, items)
		for _, s := range secrets {
			if strings.Contains(body, s) {
				t.Errorf("%s list leaks %q", label, s)
			}
		}
	}

	// Restricted address columns are not searchable by non-primary callers.
	items, _, _ := e.prop.List(ctx(), cust, usecase.ListPropertiesInput{Search: "Rahasia Pribadi", Page: 1, Limit: 10})
	if len(items) != 0 {
		t.Errorf("search matched restricted street name: %d items", len(items))
	}
}

func TestCoListingRules(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	m1 := e.user(t, entity.RoleSales, "active")
	m2 := e.user(t, entity.RoleSales, "active")
	m3 := e.user(t, entity.RoleSales, "active")
	inactive := e.user(t, entity.RoleSales, "inactive")
	cust := usecase.Actor{} // guest: no account
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, a)
	meta := usecase.RequestMeta{}

	// Non-admin cannot manage.
	for _, actor := range []usecase.Actor{a, m1} {
		if _, err := e.co.Add(ctx(), actor, id, m1.ID, meta); statusOf(err) != 403 {
			t.Errorf("non-admin add: %d", statusOf(err))
		}
		if err := e.co.Remove(ctx(), actor, id, m1.ID, meta); statusOf(err) != 403 {
			t.Errorf("non-admin remove: %d", statusOf(err))
		}
		if _, err := e.co.Replace(ctx(), actor, id, m1.ID, m2.ID, meta); statusOf(err) != 403 {
			t.Errorf("non-admin replace: %d", statusOf(err))
		}
	}

	// Ineligible users.
	if _, err := e.co.Add(ctx(), admin, id, inactive.ID, meta); statusOf(err) != 422 {
		t.Errorf("inactive: %d, want 422", statusOf(err))
	}
	if _, err := e.co.Add(ctx(), admin, id, cust.ID, meta); statusOf(err) != 422 {
		t.Errorf("non-marketing: %d, want 422", statusOf(err))
	}
	// Primary cannot be (re-)added or removed through co-listing endpoints.
	if _, err := e.co.Add(ctx(), admin, id, a.ID, meta); statusOf(err) != 409 {
		t.Errorf("add primary as co: %d, want 409", statusOf(err))
	}
	if err := e.co.Remove(ctx(), admin, id, a.ID, meta); statusOf(err) != 422 {
		t.Errorf("remove primary: %d, want 422", statusOf(err))
	}

	// 1 primary + 2 co-listing ok; duplicate and 4th rejected.
	if _, err := e.co.Add(ctx(), admin, id, m1.ID, meta); err != nil {
		t.Fatal(err)
	}
	if _, err := e.co.Add(ctx(), admin, id, m1.ID, meta); statusOf(err) != 409 {
		t.Errorf("duplicate: %d, want 409", statusOf(err))
	}
	if _, err := e.co.Add(ctx(), admin, id, m2.ID, meta); err != nil {
		t.Fatal(err)
	}
	if _, err := e.co.Add(ctx(), admin, id, m3.ID, meta); statusOf(err) != 422 {
		t.Errorf("fourth marketing: %d, want 422", statusOf(err))
	}

	// Replace and remove.
	team, err := e.co.Replace(ctx(), admin, id, m2.ID, m3.ID, meta)
	if err != nil || len(team) != 3 {
		t.Fatalf("replace: %v team=%d", err, len(team))
	}
	if err := e.co.Remove(ctx(), admin, id, m1.ID, meta); err != nil {
		t.Fatal(err)
	}

	// Co-listing grants no edit rights, and primary is unchanged.
	name := "x"
	if _, err := e.prop.Update(ctx(), m3, id, usecase.UpdatePropertyInput{Name: &name}, meta); statusOf(err) != 403 {
		t.Errorf("co-listing edit: %d, want 403", statusOf(err))
	}
	var p entity.Property
	e.db.First(&p, "id = ?", id)
	if p.PrimaryMarketingID != a.ID {
		t.Errorf("primary changed to %s", p.PrimaryMarketingID)
	}

	// Audit records exist and contain no PII.
	var logs []entity.AuditLog
	e.db.Where("module = ? AND entity_id = ?", "co_listing", id).Find(&logs)
	if len(logs) != 4 { // add m1, add m2, replace, remove (failed attempts write nothing)
		t.Errorf("audit rows = %d, want 4", len(logs))
	}
	for _, l := range logs {
		if strings.Contains(string(l.Metadata), "@") {
			t.Errorf("audit metadata may contain PII: %s", l.Metadata)
		}
	}
}

func TestCoListingConcurrencyNeverExceedsLimit(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, a)

	const n = 8
	cands := make([]usecase.Actor, n)
	for i := range cands {
		cands[i] = e.user(t, entity.RoleSales, "active")
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	ok := 0
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(c usecase.Actor) {
			defer wg.Done()
			if _, err := e.co.Add(ctx(), admin, id, c.ID, usecase.RequestMeta{}); err == nil {
				mu.Lock()
				ok++
				mu.Unlock()
			}
		}(cands[i])
	}
	wg.Wait()

	var total int64
	e.db.Model(&entity.PropertyMarketingAssignment{}).Where("property_id = ?", id).Count(&total)
	if ok != 2 || total != 3 {
		t.Fatalf("succeeded=%d total assignments=%d, want 2 and 3", ok, total)
	}
}

func TestAvailabilitySemantics(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	cust := usecase.Actor{} // guest: no account
	avail := e.listing(t, a)
	draft := e.listing(t, a)
	sold := e.listing(t, a)
	e.db.Model(&entity.Property{}).Where("id = ?", draft).Update("status", "draft")
	e.db.Model(&entity.Property{}).Where("id = ?", sold).Update("status", "sold")

	has := func(items []usecase.PublicProperty, id uuid.UUID) bool {
		for _, i := range items {
			if i.ID == id {
				return true
			}
		}
		return false
	}
	in := usecase.ListPropertiesInput{Search: "Rumah Test", Page: 1, Limit: 100, Status: "sold"} // client tries to force status
	items, _, err := e.prop.List(ctx(), cust, in)
	if err != nil {
		t.Fatal(err)
	}
	if !has(items, avail) || has(items, sold) || has(items, draft) {
		t.Errorf("guest default list wrong: avail=%v sold=%v draft=%v", has(items, avail), has(items, sold), has(items, draft))
	}
	// History section shows sold with a badge, never available/draft.
	in.History = true
	items, _, _ = e.prop.List(ctx(), cust, in)
	if !has(items, sold) || has(items, avail) || has(items, draft) {
		t.Errorf("history list wrong")
	}
	for _, i := range items {
		if i.ID == sold && i.StatusBadge != "SOLD" {
			t.Errorf("badge = %q", i.StatusBadge)
		}
	}
	// include_history returns available + sold together with is_available.
	in.History = false
	in.IncludeHistory = true
	items, _, _ = e.prop.List(ctx(), cust, in)
	for _, i := range items {
		if i.ID == sold && i.IsAvailable {
			t.Error("sold item flagged available")
		}
		if i.ID == avail && !i.IsAvailable {
			t.Error("available item flagged unavailable")
		}
	}
	if !has(items, sold) || !has(items, avail) || has(items, draft) {
		t.Errorf("include_history list wrong")
	}
	// Draft detail is hidden from guests (404) but visible to admin.
	if _, err := e.prop.GetByID(ctx(), cust, draft); statusOf(err) != 404 {
		t.Errorf("guest draft detail: %d, want 404", statusOf(err))
	}
	admin := e.user(t, entity.RoleAdmin, "active")
	if _, err := e.prop.GetByID(ctx(), admin, draft); err != nil {
		t.Errorf("admin draft detail: %v", err)
	}
	// Sold/rented cannot be set manually.
	if err := e.prop.UpdateStatus(ctx(), a, avail, "sold", usecase.RequestMeta{}); statusOf(err) != 422 {
		t.Errorf("manual sold: %d, want 422", statusOf(err))
	}
}

func TestFiltersAndWhatsApp(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	cust := usecase.Actor{} // guest: no account
	id := e.listing(t, a)

	min, max := 1_000_000_000.0, 2_000_000_000.0
	items, _, err := e.prop.List(ctx(), cust, usecase.ListPropertiesInput{Search: "Rumah Test", MinPrice: &min, MaxPrice: &max, Page: 1, Limit: 100})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, i := range items {
		found = found || i.ID == id
	}
	if !found {
		t.Error("price range filter excluded matching listing")
	}
	hi := 5_000_000_000.0
	items, _, _ = e.prop.List(ctx(), cust, usecase.ListPropertiesInput{Search: "Rumah Test", MinPrice: &hi, Page: 1, Limit: 100})
	for _, i := range items {
		if i.ID == id {
			t.Error("min_price filter failed")
		}
	}
	if _, _, err := e.prop.List(ctx(), cust, usecase.ListPropertiesInput{MinPrice: &max, MaxPrice: &min}); statusOf(err) != 422 {
		t.Error("inverted range should be 422")
	}
	if _, _, err := e.prop.List(ctx(), cust, usecase.ListPropertiesInput{Sort: "name; DROP TABLE users"}); statusOf(err) != 422 {
		t.Error("bad sort should be 422")
	}

	// WhatsApp: marketing number only, encoded, no owner phone, anonymous gets no contact.
	d, _ := e.prop.GetByID(ctx(), cust, id)
	pub := d.(usecase.PublicProperty)
	if pub.Contact == nil || pub.Contact.WhatsAppNumber != "6281200000099" {
		t.Fatalf("contact = %+v", pub.Contact)
	}
	if !strings.HasPrefix(pub.Contact.WhatsAppURL, "https://wa.me/6281200000099?text=Hello%2C%20I%20am%20interested") ||
		strings.Contains(pub.Contact.WhatsAppURL, " ") || strings.Contains(pub.Contact.WhatsAppURL, "811") {
		t.Errorf("bad url %s", pub.Contact.WhatsAppURL)
	}
	d, _ = e.prop.GetByID(ctx(), usecase.Actor{}, id)
	if c := d.(usecase.PublicProperty).Contact; c == nil || c.WhatsAppURL == "" {
		t.Error("anonymous should receive the marketing WhatsApp contact")
	}
	if strings.Contains(strings.ToLower(mustJSON(t, d)), "email") {
		t.Error("property contract must not contain an email contact action")
	}
}

func TestDealPropertyIsArchivedNotDeleted(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	id := e.listing(t, a)

	typ, price, buyer := "sale", 1.0, "Pembeli"
	deal := &entity.Deal{PropertyID: id, SalesID: a.ID, RecordedBy: a.ID, Outcome: "deal", Type: &typ, Price: &price, BuyerName: &buyer, DealDate: time.Now()}
	if err := e.db.Create(deal).Error; err != nil {
		t.Fatalf("cannot create deal fixture: %v", err)
	}
	res, err := e.prop.Delete(ctx(), a, id, usecase.RequestMeta{})
	if err != nil || !res.Archived {
		t.Fatalf("delete with deal: %v %+v", err, res)
	}
	var p entity.Property
	if err := e.db.First(&p, "id = ?", id).Error; err != nil || p.Status != "archived" {
		t.Fatalf("expected archived row, got %v %q", err, p.Status)
	}
}

var _ = fmt.Sprintf

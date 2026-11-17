package usecase_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/usecase"
)

func TestTransferOwnershipAndCoListingStatus(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	b := e.user(t, entity.RoleSales, "active")
	c := e.user(t, entity.RoleSales, "active")
	d := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	cust := usecase.Actor{} // guest: no account
	id := e.listing(t, a)
	m := usecase.RequestMeta{}

	// Co-listing may toggle availability but nothing else.
	if _, err := e.co.Add(ctx(), admin, id, b.ID, m); err != nil {
		t.Fatal(err)
	}
	if err := e.prop.UpdateStatus(ctx(), b, id, "reserved", m); err != nil {
		t.Errorf("co-listing reserve: %v", err)
	}
	for _, st := range []string{"archived", "draft", "sold", "rented"} {
		if err := e.prop.UpdateStatus(ctx(), b, id, st, m); err == nil {
			t.Errorf("co-listing set %s", st)
		}
	}
	if err := e.prop.UpdateStatus(ctx(), c, id, "available", m); statusOf(err) != 403 {
		t.Errorf("unrelated marketing status: %d, want 403", statusOf(err))
	}
	e.db.Model(&entity.Property{}).Where("id = ?", id).Update("status", "draft")
	if err := e.prop.UpdateStatus(ctx(), b, id, "available", m); statusOf(err) != 403 {
		t.Errorf("co-listing publishing a draft: %d, want 403", statusOf(err))
	}
	e.db.Model(&entity.Property{}).Where("id = ?", id).Update("status", "available")

	// Only admin can transfer.
	for _, actor := range []usecase.Actor{a, b} {
		if _, err := e.co.TransferPrimary(ctx(), actor, id, b.ID, m); statusOf(err) != 403 {
			t.Errorf("non-admin transfer: %d, want 403", statusOf(err))
		}
	}
	if _, err := e.co.TransferPrimary(ctx(), admin, id, a.ID, m); statusOf(err) != 422 {
		t.Errorf("transfer to current primary: %d, want 422", statusOf(err))
	}
	if _, err := e.co.TransferPrimary(ctx(), admin, id, cust.ID, m); statusOf(err) != 422 {
		t.Errorf("transfer to non-marketing: %d, want 422", statusOf(err))
	}

	// Transfer to an existing co-listing member: roles swap.
	team, err := e.co.TransferPrimary(ctx(), admin, id, b.ID, m)
	if err != nil || len(team) != 2 {
		t.Fatalf("transfer: %v team=%d", err, len(team))
	}
	var p entity.Property
	e.db.First(&p, "id = ?", id)
	if p.PrimaryMarketingID != b.ID {
		t.Fatalf("primary = %s, want %s", p.PrimaryMarketingID, b.ID)
	}
	var owner entity.Owner
	e.db.First(&owner, "id = ?", p.OwnerID)
	if owner.CreatedBy == nil || *owner.CreatedBy != b.ID {
		t.Error("owner record should follow the new primary")
	}
	// New primary reads owner; old primary is now read-only co-listing without owner access.
	if _, err := e.own.GetByProperty(ctx(), b, id); err != nil {
		t.Errorf("new primary owner read: %v", err)
	}
	if _, err := e.own.GetByProperty(ctx(), a, id); statusOf(err) != 403 {
		t.Errorf("old primary owner read: %d, want 403", statusOf(err))
	}
	name := "x"
	if _, err := e.prop.Update(ctx(), a, id, usecase.UpdatePropertyInput{Name: &name}, m); statusOf(err) != 403 {
		t.Errorf("old primary edit: %d, want 403", statusOf(err))
	}

	// Transfer to someone outside the team needs a free co-listing slot.
	if _, err := e.co.TransferPrimary(ctx(), admin, id, c.ID, m); err != nil {
		t.Fatalf("transfer to outsider with free slot: %v", err)
	}
	// Team now: c primary, a and b co-listing (full). Another outsider is rejected.
	if _, err := e.co.TransferPrimary(ctx(), admin, id, d.ID, m); statusOf(err) != 422 {
		t.Errorf("transfer with full team: %d, want 422", statusOf(err))
	}
	var primaries, total int64
	e.db.Model(&entity.PropertyMarketingAssignment{}).Where("property_id = ? AND assignment_type = 'primary'", id).Count(&primaries)
	e.db.Model(&entity.PropertyMarketingAssignment{}).Where("property_id = ?", id).Count(&total)
	if primaries != 1 || total != 3 {
		t.Errorf("assignments primary=%d total=%d, want 1 and 3", primaries, total)
	}
	var logs int64
	e.db.Model(&entity.AuditLog{}).Where("action = ? AND entity_id = ?", "transferred_primary_ownership", id).Count(&logs)
	if logs != 2 {
		t.Errorf("transfer audit rows = %d, want 2", logs)
	}
}

func TestFacilityFilter(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	cust := usecase.Actor{} // guest: no account
	f1 := &entity.Facility{Name: "F1 " + a.ID.String()[:8]}
	f2 := &entity.Facility{Name: "F2 " + a.ID.String()[:8]}
	e.db.Create(f1)
	e.db.Create(f2)
	price := 1.0
	mk := func(code string, ids ...uuid.UUID) uuid.UUID {
		p, err := e.prop.Create(ctx(), a, usecase.CreatePropertyInput{
			Owner: usecase.OwnerInput{Name: "Own"}, Code: code + a.ID.String()[:8], Name: "Fasilitas " + a.ID.String()[:8], Type: "house",
			TransactionType: "sale", Price: &price, FacilityIDs: ids,
		}, usecase.RequestMeta{})
		if err != nil {
			t.Fatal(err)
		}
		return p.ID
	}
	both := mk("FB-", f1.ID, f2.ID)
	only1 := mk("F1-", f1.ID)
	mk("F0-")

	ids := func(f []uuid.UUID) map[uuid.UUID]bool {
		items, _, err := e.prop.List(ctx(), cust, usecase.ListPropertiesInput{Search: "Fasilitas " + a.ID.String()[:8], FacilityIDs: f, Page: 1, Limit: 50})
		if err != nil {
			t.Fatal(err)
		}
		m := map[uuid.UUID]bool{}
		for _, i := range items {
			m[i.ID] = true
		}
		return m
	}
	if got := ids([]uuid.UUID{f1.ID}); !got[both] || !got[only1] || len(got) != 2 {
		t.Errorf("filter f1: %v", got)
	}
	if got := ids([]uuid.UUID{f1.ID, f2.ID}); !got[both] || got[only1] || len(got) != 1 {
		t.Errorf("filter f1+f2 (must match ALL): %v", got)
	}
}

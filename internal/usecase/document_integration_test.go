package usecase_test

import (
	"bytes"
	"mime/multipart"
	"net/textproto"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	pgRepo "github.com/leventa-backend/internal/repository/postgres"
	"github.com/leventa-backend/internal/storage"
	"github.com/leventa-backend/internal/usecase"
)

func (e *env) docUC(t *testing.T) (*usecase.DocumentUsecase, *storage.LocalStorage) {
	t.Helper()
	st := storage.NewDocumentStorage(t.TempDir(), 1<<20)
	uc := usecase.NewDocumentUsecase(pgRepo.NewPropertyRepository(e.db), pgRepo.NewCoListingRepository(e.db),
		pgRepo.NewPropertyDocumentRepository(e.db), pgRepo.NewTransactor(e.db), st, e.audit())
	return uc, st
}

var pdfBytes = []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF\n")

// upload builds a multipart file the way the HTTP layer would.
func upload(t *testing.T, name string, content []byte) (multipart.File, *multipart.FileHeader) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+name+`"`)
	part, err := w.CreatePart(h)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write(content)
	_ = w.Close()
	form, err := multipart.NewReader(&buf, w.Boundary()).ReadForm(1 << 20)
	if err != nil {
		t.Fatal(err)
	}
	fh := form.File["file"][0]
	f, err := fh.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	return f, fh
}

func TestOwnershipDocuments(t *testing.T) {
	e := setup(t)
	uc, st := e.docUC(t)
	primary := e.user(t, entity.RoleSales, "active")
	co := e.user(t, entity.RoleSales, "active")
	other := e.user(t, entity.RoleSales, "active")
	admin := e.user(t, entity.RoleAdmin, "active")
	id := e.listing(t, primary) // house, no certificate flags yet
	if _, err := e.co.Add(ctx(), admin, id, co.ID, usecase.RequestMeta{}); err != nil {
		t.Fatal(err)
	}
	m := usecase.RequestMeta{}
	up := func(actor usecase.Actor, docType string) (*entity.PropertyDocument, error) {
		f, h := upload(t, "sertifikat.pdf", pdfBytes)
		return uc.Upload(ctx(), actor, id, docType, f, h, m)
	}

	// The flag must be true first.
	if _, err := up(primary, "shm"); statusOf(err) != 422 {
		t.Errorf("upload with cert_shm=false: %d, want 422", statusOf(err))
	}
	yes := true
	if _, err := e.prop.Update(ctx(), primary, id, usecase.UpdatePropertyInput{PropertyDetailsPatch: usecase.PropertyDetailsPatch{CertSHM: &yes}}, m); err != nil {
		t.Fatal(err)
	}
	// A document type that does not apply to a house is refused even when asked.
	if _, err := up(primary, "shmsrs"); statusOf(err) != 422 {
		t.Errorf("shmsrs on a house: %d, want 422", statusOf(err))
	}
	// Only the primary marketing uploads.
	for label, actor := range map[string]usecase.Actor{"co-listing": co, "other marketing": other, "admin": admin} {
		if _, err := up(actor, "shm"); statusOf(err) != 403 {
			t.Errorf("%s upload: %d, want 403", label, statusOf(err))
		}
	}
	if _, err := up(usecase.Actor{}, "shm"); statusOf(err) != 401 {
		t.Errorf("guest upload: %d, want 401", statusOf(err))
	}

	doc, err := up(primary, "shm")
	if err != nil {
		t.Fatal(err)
	}
	if doc.MimeType != "application/pdf" || doc.OriginalName != "sertifikat.pdf" {
		t.Errorf("document metadata wrong: %+v", doc)
	}
	if b := mustJSON(t, doc); bytes.Contains([]byte(b), []byte("property-documents")) {
		t.Error("storage path leaked in JSON")
	}

	// Non-document content is refused by content sniffing, whatever the extension says.
	f, h := upload(t, "evil.pdf", []byte("<html><script>alert(1)</script></html>"))
	if _, err := uc.Upload(ctx(), primary, id, "shm", f, h, m); statusOf(err) != 400 {
		t.Errorf("html disguised as pdf: %d, want 400", statusOf(err))
	}

	// Read access: primary and co-listing; nobody else.
	for label, actor := range map[string]usecase.Actor{"primary": primary, "co-listing": co} {
		list, err := uc.List(ctx(), actor, id)
		if err != nil || len(list) != 1 {
			t.Errorf("%s list: %v %d", label, err, len(list))
		}
		file, err := uc.Open(ctx(), actor, id, doc.ID, m)
		if err != nil {
			t.Fatalf("%s open: %v", label, err)
		}
		got, rerr := os.ReadFile(file.Path)
		if rerr != nil || !bytes.Equal(got, pdfBytes) {
			t.Errorf("%s downloaded wrong content: %v", label, rerr)
		}
	}
	for label, actor := range map[string]usecase.Actor{"other marketing": other, "admin": admin} {
		if _, err := uc.List(ctx(), actor, id); statusOf(err) != 403 {
			t.Errorf("%s list: %d, want 403", label, statusOf(err))
		}
		if _, err := uc.Open(ctx(), actor, id, doc.ID, m); statusOf(err) != 403 {
			t.Errorf("%s download: %d, want 403", label, statusOf(err))
		}
	}
	if _, err := uc.Open(ctx(), usecase.Actor{}, id, doc.ID, m); statusOf(err) != 401 {
		t.Errorf("guest download: %d, want 401", statusOf(err))
	}
	// IDOR: a document id of another listing cannot be fetched through this one.
	otherListing := e.listing(t, primary)
	if _, err := uc.Open(ctx(), primary, otherListing, doc.ID, m); statusOf(err) != 404 {
		t.Errorf("cross-listing download: %d, want 404", statusOf(err))
	}

	// Re-upload replaces the document and removes the old file.
	again, err := up(primary, "shm")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uc.Open(ctx(), primary, id, doc.ID, m); statusOf(err) != 404 {
		t.Errorf("replaced document still reachable: %d", statusOf(err))
	}
	if list, _ := uc.List(ctx(), primary, id); len(list) != 1 || list[0].ID != again.ID {
		t.Errorf("replace left %d documents", len(list))
	}

	// The flag cannot be cleared while the document exists.
	no := false
	if _, err := e.prop.Update(ctx(), primary, id, usecase.UpdatePropertyInput{PropertyDetailsPatch: usecase.PropertyDetailsPatch{CertSHM: &no}}, m); statusOf(err) != 409 {
		t.Errorf("clear flag with document: %d, want 409", statusOf(err))
	}
	if err := uc.Delete(ctx(), co, id, again.ID, m); statusOf(err) != 403 {
		t.Errorf("co-listing delete: %d, want 403", statusOf(err))
	}
	if err := uc.Delete(ctx(), primary, id, again.ID, m); err != nil {
		t.Fatal(err)
	}
	if _, err := e.prop.Update(ctx(), primary, id, usecase.UpdatePropertyInput{PropertyDetailsPatch: usecase.PropertyDetailsPatch{CertSHM: &no}}, m); err != nil {
		t.Errorf("clear flag after delete: %v", err)
	}
	_ = st
}

func TestCertificateFlagsMatchPropertyType(t *testing.T) {
	e := setup(t)
	a := e.user(t, entity.RoleSales, "active")
	price := 1.0
	base := func(typ string) usecase.CreatePropertyInput {
		return usecase.CreatePropertyInput{
			Owner: usecase.OwnerInput{Name: "Own"}, Code: "C-" + uuid.NewString()[:12], Name: "Properti Sertifikat",
			Type: typ, TransactionType: "sale", Price: &price,
		}
	}
	// SHMSRS (apartment unit title) is not valid for a house; HGU is factory-only.
	in := base("house")
	in.CertSHMSRS = true
	if _, err := e.prop.Create(ctx(), a, in, usecase.RequestMeta{}); statusOf(err) != 422 {
		t.Errorf("house with cert_shmsrs: %d, want 422", statusOf(err))
	}
	in = base("apartment")
	in.CertSHMSRS = true
	if _, err := e.prop.Create(ctx(), a, in, usecase.RequestMeta{}); err != nil {
		t.Errorf("apartment with cert_shmsrs: %v", err)
	}
	in = base("shop_house")
	in.CertHGU = true
	if _, err := e.prop.Create(ctx(), a, in, usecase.RequestMeta{}); statusOf(err) != 422 {
		t.Errorf("shop with cert_hgu: %d, want 422", statusOf(err))
	}
	in = base("factory")
	in.CertHGU, in.CertSLF = true, true
	if _, err := e.prop.Create(ctx(), a, in, usecase.RequestMeta{}); err != nil {
		t.Errorf("factory with hgu+slf: %v", err)
	}
}

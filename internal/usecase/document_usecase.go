package usecase

import (
	"context"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/leventa-backend/internal/domain/entity"
	"github.com/leventa-backend/internal/domain/repository"
	"github.com/leventa-backend/internal/service"
	"github.com/leventa-backend/pkg/errs"
)

// DocumentStorage stores ownership documents privately (no public URL).
type DocumentStorage interface {
	Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error)
	Delete(ctx context.Context, path string) error
	Path(rel string) (string, error)
}

// DocumentUsecase manages ownership documents (SHM, HGB, ...). A document can be
// stored only while the matching cert_* flag of the property is true. Only the
// primary marketing owner uploads or deletes; primary and co-listing marketing
// read and download. Admin and guests have no access (D-029).
type DocumentUsecase struct {
	propertyAccess
	docRepo      repository.PropertyDocumentRepository
	tx           repository.Transactor
	storage      DocumentStorage
	auditService *service.AuditService
}

// NewDocumentUsecase creates a document usecase.
func NewDocumentUsecase(
	propertyRepo repository.PropertyRepository,
	coRepo repository.CoListingRepository,
	docRepo repository.PropertyDocumentRepository,
	tx repository.Transactor,
	storage DocumentStorage,
	auditService *service.AuditService,
) *DocumentUsecase {
	return &DocumentUsecase{
		propertyAccess: propertyAccess{propertyRepo: propertyRepo, coRepo: coRepo},
		docRepo:        docRepo,
		tx:             tx,
		storage:        storage,
		auditService:   auditService,
	}
}

// DocumentFile is a stored file ready to be sent to the caller.
type DocumentFile struct {
	Path     string
	Name     string
	MimeType string
}

// canRead returns the property when the actor is primary or co-listing marketing.
func (uc *DocumentUsecase) canRead(ctx context.Context, actor Actor, propertyID uuid.UUID) (*entity.Property, error) {
	if !actor.Authenticated() {
		return nil, errs.ErrUnauthorized
	}
	p, lvl, err := uc.load(ctx, actor, propertyID)
	if err != nil {
		return nil, err
	}
	if lvl != AccessPrimary && lvl != AccessCoListing {
		return nil, errs.Forbidden("Only the primary or a co-listing marketing user can access ownership documents")
	}
	return p, nil
}

// Upload stores (or replaces) the document of docType. Primary marketing only.
func (uc *DocumentUsecase) Upload(ctx context.Context, actor Actor, propertyID uuid.UUID, docType string, file multipart.File, header *multipart.FileHeader, meta RequestMeta) (*entity.PropertyDocument, error) {
	p, err := uc.requirePrimary(ctx, actor, propertyID)
	if err != nil {
		return nil, err
	}
	if !entity.DocTypeAllowed(p.Type, docType) {
		return nil, errs.Validation("document type " + docType + " does not apply to a " + p.Type + " listing")
	}
	if !p.CertFlag(docType) {
		return nil, errs.Validation("set cert_" + docType + " to true on the property before uploading this document")
	}

	path, err := uc.storage.Upload(ctx, file, header)
	if err != nil {
		return nil, errs.BadRequest(err.Error())
	}
	doc := &entity.PropertyDocument{
		PropertyID:   propertyID,
		DocType:      docType,
		StoragePath:  path,
		OriginalName: sanitizeFilename(header.Filename),
		MimeType:     mimeFromPath(path),
		SizeBytes:    header.Size,
		UploadedBy:   actor.ID,
	}

	var old *entity.PropertyDocument
	err = uc.tx.Do(ctx, func(ctx context.Context) error {
		if err := uc.propertyRepo.LockByID(ctx, propertyID); err != nil {
			return err
		}
		var err error
		if old, err = uc.docRepo.FindByType(ctx, propertyID, docType); err != nil {
			return err
		}
		if old != nil {
			if err := uc.docRepo.Delete(ctx, old.ID); err != nil {
				return err
			}
		}
		if err := uc.docRepo.Create(ctx, doc); err != nil {
			return err
		}
		return uc.auditService.LogSync(ctx, &actor.ID, "uploaded_property_document", "property_documents", &doc.ID,
			map[string]string{"property_id": propertyID.String(), "doc_type": docType}, meta.IP, meta.UserAgent)
	})
	if err != nil {
		_ = uc.storage.Delete(ctx, path)
		return nil, mapTxError(err)
	}
	if old != nil {
		_ = uc.storage.Delete(ctx, old.StoragePath)
	}
	return doc, nil
}

// List returns document metadata (never file paths). Primary and co-listing marketing only.
func (uc *DocumentUsecase) List(ctx context.Context, actor Actor, propertyID uuid.UUID) ([]entity.PropertyDocument, error) {
	if _, err := uc.canRead(ctx, actor, propertyID); err != nil {
		return nil, err
	}
	list, err := uc.docRepo.ListByProperty(ctx, propertyID)
	if err != nil {
		return nil, errs.Internal(err)
	}
	return list, nil
}

// Open returns the file to send. Primary and co-listing marketing only; every
// download is audited.
func (uc *DocumentUsecase) Open(ctx context.Context, actor Actor, propertyID, docID uuid.UUID, meta RequestMeta) (*DocumentFile, error) {
	if _, err := uc.canRead(ctx, actor, propertyID); err != nil {
		return nil, err
	}
	doc, err := uc.docRepo.FindByID(ctx, docID)
	if err != nil {
		return nil, err
	}
	if doc.PropertyID != propertyID {
		return nil, errs.NotFound("Document")
	}
	full, err := uc.storage.Path(doc.StoragePath)
	if err != nil {
		return nil, errs.Internal(err)
	}
	uc.auditService.Log(&actor.ID, "downloaded_property_document", "property_documents", &doc.ID,
		map[string]string{"property_id": propertyID.String(), "doc_type": doc.DocType}, meta.IP, meta.UserAgent)
	return &DocumentFile{Path: full, Name: doc.OriginalName, MimeType: doc.MimeType}, nil
}

// Delete removes a document. Primary marketing only.
func (uc *DocumentUsecase) Delete(ctx context.Context, actor Actor, propertyID, docID uuid.UUID, meta RequestMeta) error {
	if _, err := uc.requirePrimary(ctx, actor, propertyID); err != nil {
		return err
	}
	doc, err := uc.docRepo.FindByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc.PropertyID != propertyID {
		return errs.NotFound("Document")
	}
	if err := uc.docRepo.Delete(ctx, docID); err != nil {
		return errs.Internal(err)
	}
	_ = uc.storage.Delete(ctx, doc.StoragePath)
	uc.auditService.Log(&actor.ID, "deleted_property_document", "property_documents", &doc.ID,
		map[string]string{"property_id": propertyID.String(), "doc_type": doc.DocType}, meta.IP, meta.UserAgent)
	return nil
}

// hasDocument reports whether a document of docType is stored for the property.
func (uc *DocumentUsecase) hasDocument(ctx context.Context, propertyID uuid.UUID, docType string) (bool, error) {
	d, err := uc.docRepo.FindByType(ctx, propertyID, docType)
	return d != nil, err
}

func sanitizeFilename(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.Map(func(r rune) rune {
		if r < 32 || r == '"' || r == 127 {
			return -1
		}
		return r
	}, name)
	if name == "" || name == "." || name == "/" {
		return "document"
	}
	if len(name) > 200 {
		name = name[len(name)-200:]
	}
	return name
}

func mimeFromPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".pdf":
		return "application/pdf"
	case ".png":
		return "image/png"
	default:
		return "image/jpeg"
	}
}

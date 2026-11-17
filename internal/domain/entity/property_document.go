package entity

import (
	"time"

	"github.com/google/uuid"
)

// Ownership document types (D-029).
const (
	DocTypeSHM    = "shm"
	DocTypeHGB    = "hgb"
	DocTypeHGU    = "hgu"
	DocTypeSHMSRS = "shmsrs" // apartment unit title
	DocTypeIMB    = "imb"    // IMB / PBG
	DocTypePBB    = "pbb"
	DocTypeGirik  = "girik"
	DocTypeSLF    = "slf"
	DocTypeOther  = "other"
)

// allowedDocTypes lists which ownership documents apply to each property type.
var allowedDocTypes = map[string][]string{
	PropertyTypeHouse:     {DocTypeSHM, DocTypeHGB, DocTypeIMB, DocTypePBB, DocTypeGirik, DocTypeOther},
	PropertyTypeApartment: {DocTypeSHMSRS, DocTypeHGB, DocTypeIMB, DocTypePBB, DocTypeSLF, DocTypeOther},
	PropertyTypeShopHouse: {DocTypeSHM, DocTypeHGB, DocTypeIMB, DocTypePBB, DocTypeSLF, DocTypeOther},
	PropertyTypeFactory:   {DocTypeSHM, DocTypeHGB, DocTypeHGU, DocTypeIMB, DocTypePBB, DocTypeSLF, DocTypeOther},
}

// AllowedDocTypes returns the document types that apply to a property type.
func AllowedDocTypes(propertyType string) []string {
	return allowedDocTypes[propertyType]
}

// DocTypeAllowed reports whether docType applies to propertyType.
func DocTypeAllowed(propertyType, docType string) bool {
	for _, t := range allowedDocTypes[propertyType] {
		if t == docType {
			return true
		}
	}
	return false
}

// ValidDocTypes returns every document type.
func ValidDocTypes() []string {
	return []string{DocTypeSHM, DocTypeHGB, DocTypeHGU, DocTypeSHMSRS, DocTypeIMB, DocTypePBB, DocTypeGirik, DocTypeSLF, DocTypeOther}
}

// CertFlag reports the cert_* flag that matches docType.
func (p *Property) CertFlag(docType string) bool {
	switch docType {
	case DocTypeSHM:
		return p.CertSHM
	case DocTypeHGB:
		return p.CertHGB
	case DocTypeHGU:
		return p.CertHGU
	case DocTypeSHMSRS:
		return p.CertSHMSRS
	case DocTypeIMB:
		return p.CertIMB
	case DocTypePBB:
		return p.CertPBB
	case DocTypeGirik:
		return p.CertGirik
	case DocTypeSLF:
		return p.CertSLF
	case DocTypeOther:
		return p.CertOther
	}
	return false
}

// PropertyDocument is a stored ownership document. The file itself is served
// only through the authorized download endpoint; StoragePath is never serialized.
type PropertyDocument struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PropertyID   uuid.UUID `gorm:"type:uuid;not null;index" json:"property_id"`
	DocType      string    `gorm:"size:20;not null" json:"doc_type"`
	StoragePath  string    `gorm:"size:500;not null" json:"-"`
	OriginalName string    `gorm:"size:255;not null" json:"original_name"`
	MimeType     string    `gorm:"size:100;not null" json:"mime_type"`
	SizeBytes    int64     `gorm:"not null" json:"size_bytes"`
	UploadedBy   uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName returns the table name for PropertyDocument.
func (PropertyDocument) TableName() string { return "property_documents" }

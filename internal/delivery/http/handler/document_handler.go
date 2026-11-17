package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/leventa-backend/internal/usecase"
	"github.com/leventa-backend/pkg/response"
	"github.com/leventa-backend/pkg/validator"
)

// DocumentHandler serves ownership documents of a property (private files).
type DocumentHandler struct {
	uc        *usecase.DocumentUsecase
	validator *validator.Validator
}

// NewDocumentHandler creates a document handler.
func NewDocumentHandler(uc *usecase.DocumentUsecase, v *validator.Validator) *DocumentHandler {
	return &DocumentHandler{uc: uc, validator: v}
}

type docTypeForm struct {
	DocType string `form:"doc_type" validate:"required,oneof=shm hgb hgu shmsrs imb pbb girik slf other"`
}

// Upload stores a document (multipart: doc_type + file). Primary marketing only.
func (h *DocumentHandler) Upload(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	form := docTypeForm{DocType: c.FormValue("doc_type")}
	if errs := h.validator.Validate(form); errs != nil {
		return response.Error(c, fiber.StatusUnprocessableEntity, "Validation failed", errs)
	}
	file, err := c.FormFile("file")
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "File is required", nil)
	}
	f, err := file.Open()
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Failed to read file", nil)
	}
	defer f.Close()

	doc, err := h.uc.Upload(c.Context(), actorFrom(c), propertyID, form.DocType, f, file, metaFrom(c))
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusCreated, "Document uploaded", doc)
}

// List returns document metadata (primary and co-listing marketing).
func (h *DocumentHandler) List(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	docs, err := h.uc.List(c.Context(), actorFrom(c), propertyID)
	if err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Documents retrieved", docs)
}

// Download sends the file as an attachment (primary and co-listing marketing).
func (h *DocumentHandler) Download(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	docID, err := uuid.Parse(c.Params("docId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid document ID", nil)
	}
	f, err := h.uc.Open(c.Context(), actorFrom(c), propertyID, docID, metaFrom(c))
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderContentType, f.MimeType)
	c.Set("X-Content-Type-Options", "nosniff")
	c.Set(fiber.HeaderCacheControl, "private, no-store")
	return c.Download(f.Path, f.Name)
}

// Delete removes a document (primary marketing only).
func (h *DocumentHandler) Delete(c fiber.Ctx) error {
	propertyID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid property ID", nil)
	}
	docID, err := uuid.Parse(c.Params("docId"))
	if err != nil {
		return response.Error(c, fiber.StatusBadRequest, "Invalid document ID", nil)
	}
	if err := h.uc.Delete(c.Context(), actorFrom(c), propertyID, docID, metaFrom(c)); err != nil {
		return err
	}
	return response.Success(c, fiber.StatusOK, "Document deleted", nil)
}

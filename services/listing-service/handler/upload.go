package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/vexyruu/LIP/listing-service/model"
	"github.com/vexyruu/LIP/listing-service/store"
	"github.com/vexyruu/LIP/shared/storage"
)

func (h *Handler) CreateUpload(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if h.Storage == nil {
		http.Error(w, "upload storage is not configured", http.StatusServiceUnavailable)
		return
	}

	ext, err := storage.ExtensionForContentType(req.ContentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	uploadID := ""
	objectKey := ""
	publicURL := ""
	expiresAt := time.Now().UTC().Add(storage.PresignTTL)

	id, err := newUploadID()
	if err != nil {
		http.Error(w, "failed to create upload", http.StatusInternalServerError)
		return
	}
	objectKey = "listings/" + req.UserID + "/" + id + ext
	publicURL = h.Storage.PublicURL(objectKey)
	uploadID = id

	if err := store.CreateUpload(r.Context(), h.Pool, uploadID, &store.UploadRecord{
		UserID:           req.UserID,
		ObjectKey:        objectKey,
		PublicURL:        publicURL,
		ContentType:      req.ContentType,
		SizeBytes:        req.SizeBytes,
		ExpiresAt:        expiresAt,
		OriginalFilename: req.Filename,
	}); err != nil {
		http.Error(w, "failed to create upload", http.StatusInternalServerError)
		return
	}

	uploadURL, err := h.Storage.PresignPut(r.Context(), objectKey, req.ContentType)
	if err != nil {
		http.Error(w, "failed to presign upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.CreateUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		PublicURL: publicURL,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func (h *Handler) CompleteUpload(w http.ResponseWriter, r *http.Request) {
	if h.Storage == nil {
		http.Error(w, "upload storage is not configured", http.StatusServiceUnavailable)
		return
	}

	uploadID := r.PathValue("id")
	var req model.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rec, err := store.GetUpload(r.Context(), h.Pool, uploadID)
	if err != nil {
		if errors.Is(err, store.ErrUploadNotFound) {
			http.Error(w, "upload not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to get upload", http.StatusInternalServerError)
		return
	}
	if rec.UserID != req.UserID {
		http.Error(w, "upload does not belong to user", http.StatusForbidden)
		return
	}
	if rec.Status == "READY" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model.CompleteUploadResponse{
			UploadID:  rec.ID,
			PublicURL: rec.PublicURL,
			Status:    rec.Status,
		})
		return
	}
	if rec.Status != "PENDING" {
		http.Error(w, "upload is not pending", http.StatusConflict)
		return
	}
	if time.Now().UTC().After(rec.ExpiresAt) {
		http.Error(w, "upload expired", http.StatusGone)
		return
	}

	info, err := h.Storage.StatObject(r.Context(), rec.ObjectKey)
	if err != nil {
		http.Error(w, "uploaded object not found", http.StatusBadRequest)
		return
	}
	if info.Size <= 0 || info.Size > model.MaxUploadSizeBytes {
		http.Error(w, "uploaded object has invalid size", http.StatusBadRequest)
		return
	}
	if info.Size != rec.SizeBytes {
		http.Error(w, "uploaded object size does not match declared size", http.StatusBadRequest)
		return
	}

	header, err := h.Storage.ReadHeader(r.Context(), rec.ObjectKey, 16)
	if err != nil {
		http.Error(w, "failed to read uploaded object", http.StatusBadRequest)
		return
	}
	if err := storage.ValidateImageHeader(rec.ContentType, header); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := store.MarkUploadReady(r.Context(), h.Pool, uploadID); err != nil {
		http.Error(w, "failed to finalize upload", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.CompleteUploadResponse{
		UploadID:  rec.ID,
		PublicURL: rec.PublicURL,
		Status:    "READY",
	})
}

func (h *Handler) validateListingImages(r *http.Request, req *model.CreateListingRequest) error {
	for _, imageURL := range req.Images {
		if h.AllowExternalImageURLs && !storage.ManagedURL(h.UploadPublicBaseURL, imageURL) {
			continue
		}
		ok, err := store.VerifyUploadReady(r.Context(), h.Pool, req.UserID, imageURL)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("image URL is not a ready upload owned by user")
		}
	}
	return nil
}

func newUploadID() (string, error) {
	return storeNewUUID()
}

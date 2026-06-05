package model

import (
	"errors"
	"regexp"
	"strings"
)

const MaxUploadSizeBytes = 10 * 1024 * 1024

var uploadContentTypeRE = regexp.MustCompile(`^image/(jpeg|png|webp)$`)

type CreateUploadRequest struct {
	UserID      string `json:"user_id"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	SizeBytes   int64  `json:"size_bytes"`
}

func (r *CreateUploadRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if !uuidRE.MatchString(r.UserID) {
		return errors.New("user_id must be a valid UUID")
	}
	if strings.TrimSpace(r.Filename) == "" {
		return errors.New("filename is required")
	}
	if len(r.Filename) > 255 {
		return errors.New("filename must be 255 characters or fewer")
	}
	contentType := strings.ToLower(strings.TrimSpace(r.ContentType))
	if !uploadContentTypeRE.MatchString(contentType) {
		return errors.New("content_type must be image/jpeg, image/png, or image/webp")
	}
	r.ContentType = contentType
	if r.SizeBytes <= 0 {
		return errors.New("size_bytes must be greater than 0")
	}
	if r.SizeBytes > MaxUploadSizeBytes {
		return errors.New("size_bytes exceeds 10MB limit")
	}
	return nil
}

type CreateUploadResponse struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	ExpiresAt string `json:"expires_at"`
}

type CompleteUploadRequest struct {
	UserID string `json:"user_id"`
}

func (r *CompleteUploadRequest) Validate() error {
	if r.UserID == "" {
		return errors.New("user_id is required")
	}
	if !uuidRE.MatchString(r.UserID) {
		return errors.New("user_id must be a valid UUID")
	}
	return nil
}

type CompleteUploadResponse struct {
	UploadID  string `json:"upload_id"`
	PublicURL string `json:"public_url"`
	Status    string `json:"status"`
}

package storage

import "testing"

func TestValidateImageHeader(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		header      []byte
		wantErr     bool
	}{
		{
			name:        "valid jpeg",
			contentType: "image/jpeg",
			header:      []byte{0xFF, 0xD8, 0xFF, 0xE0},
		},
		{
			name:        "invalid jpeg",
			contentType: "image/jpeg",
			header:      []byte{0x89, 0x50, 0x4E, 0x47},
			wantErr:     true,
		},
		{
			name:        "valid png",
			contentType: "image/png",
			header:      []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateImageHeader(tt.contentType, tt.header)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestManagedURL(t *testing.T) {
	base := "http://localhost:9000/mlip-listings"
	if !ManagedURL(base, base+"/listings/user/file.jpg") {
		t.Fatal("expected managed URL")
	}
	if ManagedURL(base, "https://images.unsplash.com/photo.jpg") {
		t.Fatal("expected external URL to be unmanaged")
	}
}

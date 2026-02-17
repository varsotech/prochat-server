package clientmetadata

import (
	"errors"
	"testing"
)

func TestClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientId string
		wantErr  error
	}{
		{
			name:     "valid",
			clientId: "https://example.com/test",
		},
		{
			name:     "empty",
			clientId: "",
			wantErr:  errClientIDEmpty,
		},
		{
			name:     "http instead of https",
			clientId: "http://example.com/test",
			wantErr:  errClientIDMissingHTTPS,
		},
		{
			name:     "no url path component",
			clientId: "https://example.com/",
			wantErr:  errClientIDNoURLPath,
		},
		{
			name:     "no url path component: no root /",
			clientId: "https://example.com",
			wantErr:  errClientIDNoURLPath,
		},
		{
			name:     "contains single-dot segment",
			clientId: "https://example.com/hello/world/./goodbye",
			wantErr:  errClientIDNoDotPathSegment,
		},
		{
			name:     "contains double-dot segment",
			clientId: "https://example.com/hello/world/../goodbye",
			wantErr:  errClientIDNoDotPathSegment,
		},
		{
			name:     "contains path fragment",
			clientId: "https://example.com/hello/world#goodbye",
			wantErr:  errClientIDNoFragment,
		},
		{
			name:     "contains query params",
			clientId: "https://example.com/hello/world?good=bye",
			wantErr:  errClientIDNoQueryParams,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClientID(tt.clientId, true)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = '%v', wantErr '%v'", err, tt.wantErr)
			}
		})
	}
}

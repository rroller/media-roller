package utils

import "testing"

func TestNormalizeUrl(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{url: "example.com", want: "example.com"},
		{url: "https://example.com", want: "https://example.com"},
		{url: "https://example.com this is an example", want: "https://example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := NormalizeUrl(tt.url); got != tt.want {
				t.Errorf("NormalizeUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

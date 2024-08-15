package auth

import (
	"reflect"
	"testing"
)

func Test_parseRegistryUrl(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *codeArtifactSpec
		wantErr bool
	}{
		{
			name: "Valid CodeArtifact URL",
			url:  "https://envtio-prod-481471033259.d.codeartifact.us-east-1.amazonaws.com/npm/npm-releases/",
			want: &codeArtifactSpec{
				domain:  "envtio-prod",
				account: "481471033259",
				region:  "us-east-1",
			},
			wantErr: false,
		},
		{
			name:    "Invalid CodeArtifact URL",
			url:     "https://example.com",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRegistryUrl(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRegistryUrl() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRegistryUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}

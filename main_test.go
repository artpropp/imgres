package main

import (
	"reflect"
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		s       string
		want    picSize
		wantErr bool
	}{
		{"500x500", picSize{500, 500}, false},
		{"500-500", picSize{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got, err := parseSize(tt.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize(%s) error = %v, wantErr %v", tt.s, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSize(%s)=%v want %v", tt.s, got, tt.want)
			}
		})
	}
}

func TestUseFile(t *testing.T) {
	tests := []struct {
		fileName string
		want     bool
	}{
		{"a/b/file.JPG", true},
		{"file.jpeg", true},
		{"no_image.ABC", false},
		{"no_extension", false},
	}
	for _, tt := range tests {
		t.Run(tt.fileName, func(t *testing.T) {
			if got := useFile(tt.fileName); got != tt.want {
				t.Errorf("useFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

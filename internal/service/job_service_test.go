package service

import (
	"sort"
	"testing"
)

func TestCalculateMatch(t *testing.T) {
	tests := []struct {
		name        string
		user        []string
		job         []string
		wantPercent int
		wantMatched []string
		wantMissing []string
	}{
		{
			name:        "perfect match",
			user:        []string{"Go", "Docker"},
			job:         []string{"Go", "Docker"},
			wantPercent: 100,
			wantMatched: []string{"docker", "go"},
			wantMissing: nil,
		},
		{
			name:        "half match",
			user:        []string{"Go"},
			job:         []string{"Go", "Java"},
			wantPercent: 50,
			wantMatched: []string{"go"},
			wantMissing: []string{"java"},
		},
		{
			name:        "no match",
			user:        []string{"PHP"},
			job:         []string{"Rust"},
			wantPercent: 0,
			wantMatched: nil,
			wantMissing: []string{"rust"},
		},
		{
			name:        "empty user skills",
			user:        []string{},
			job:         []string{"Go"},
			wantPercent: 0,
			wantMatched: nil,
			wantMissing: []string{"go"},
		},
		{
			name:        "case insensitive",
			user:        []string{"go", "DOCKER"},
			job:         []string{"Go", "docker"},
			wantPercent: 100,
			wantMatched: []string{"docker", "go"},
			wantMissing: nil,
		},
		{
			name:        "both empty",
			user:        []string{},
			job:         []string{},
			wantPercent: 0,
			wantMatched: nil,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percent, matched, missing := CalculateMatch(tt.user, tt.job)
			if percent != tt.wantPercent {
				t.Errorf("percent: got %d, want %d", percent, tt.wantPercent)
			}
			sort.Strings(matched)
			sort.Strings(missing)
			if !equalSlices(matched, tt.wantMatched) {
				t.Errorf("matched: got %v, want %v", matched, tt.wantMatched)
			}
			if !equalSlices(missing, tt.wantMissing) {
				t.Errorf("missing: got %v, want %v", missing, tt.wantMissing)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

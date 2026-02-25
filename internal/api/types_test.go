package api

import (
	"testing"
)

func TestParseOutcomePrices(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "valid prices",
			raw:  `["0.65","0.35"]`,
			want: []string{"0.65", "0.35"},
		},
		{
			name: "empty string",
			raw:  "",
			want: nil,
		},
		{
			name:    "invalid json",
			raw:     "not json",
			wantErr: true,
		},
		{
			name: "three outcomes",
			raw:  `["0.50","0.30","0.20"]`,
			want: []string{"0.50", "0.30", "0.20"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOutcomePrices(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOutcomePrices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ParseOutcomePrices() = %v, want %v", got, tt.want)
					return
				}
				for i := range got {
					if got[i] != tt.want[i] {
						t.Errorf("ParseOutcomePrices()[%d] = %q, want %q", i, got[i], tt.want[i])
					}
				}
			}
		})
	}
}

func TestParseOutcomes(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "yes/no",
			raw:  `["Yes","No"]`,
			want: []string{"Yes", "No"},
		},
		{
			name: "empty",
			raw:  "",
			want: nil,
		},
		{
			name: "multiple",
			raw:  `["Red","Blue","Green"]`,
			want: []string{"Red", "Blue", "Green"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseOutcomes(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseOutcomes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ParseOutcomes() = %v, want %v", got, tt.want)
			}
		})
	}
}

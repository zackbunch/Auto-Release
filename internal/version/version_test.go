package version

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      Version
		expectErr bool
	}{
		{
			name:  "Valid version",
			input: "1.2.3",
			want:  Version{Major: 1, Minor: 2, Patch: 3},
		},
		{
			name:      "Missing patch",
			input:     "1.2",
			expectErr: true,
		},
		{
			name:      "Too many segments",
			input:     "1.2.3.4",
			expectErr: true,
		},
		{
			name:      "Non-numeric parts",
			input:     "a.b.c",
			expectErr: true,
		},
		{
			name:      "Invalid major version",
			input:     "a.2.3",
			expectErr: true,
		},
		{
			name:      "Invalid minor version",
			input:     "1.b.3",
			expectErr: true,
		},
		{
			name:      "Invalid patch version",
			input:     "1.2.c",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if got != tt.want {
					t.Errorf("expected %+v, got %+v", tt.want, got)
				}
			}
		})
	}
}

func TestInc(t *testing.T) {
	base := Version{Major: 1, Minor: 2, Patch: 3}

	tests := []struct {
		name string
		bump VersionType
		want Version
	}{
		{
			name: "Patch bump",
			bump: Patch,
			want: Version{Major: 1, Minor: 2, Patch: 4},
		},
		{
			name: "Minor bump",
			bump: Minor,
			want: Version{Major: 1, Minor: 3, Patch: 0},
		},
		{
			name: "Major bump",
			bump: Major,
			want: Version{Major: 2, Minor: 0, Patch: 0},
		},
		{
			name: "Unknown bump (no change)",
			bump: "unknown",
			want: base,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := base.Increment(tt.bump)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %+v, got %+v", tt.want, got)
			}
		})
	}
}

func TestString(t *testing.T) {
	v := Version{Major: 1, Minor: 2, Patch: 3}
	got := v.String()
	want := "1.2.3"

	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

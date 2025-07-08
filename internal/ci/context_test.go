package ci

import (
	"os"
	"testing"
)

func TestLoadContext(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		config   string // YAML config content
		expected Context
		expectErr bool
	}{
		{
			name: "Feature Branch",
			env: map[string]string{
				"CI_COMMIT_REF_NAME": "feature/new-thing",
				"CI_COMMIT_SHA":      "abc1234",
				"CI_PROJECT_PATH":    "my-group/my-project",
				"CI_REGISTRY_IMAGE":  "registry.example.com/my-group/my-project",
				"CI_DEFAULT_BRANCH":  "main",
			},
			config: `protected_branches: ["main", "dev"]`,
			expected: Context{
				RefName:         "feature/new-thing",
				SHA:             "abc1234",
				ProjectPath:     "my-group/my-project",
				RegistryImage:   "registry.example.com/my-group/my-project",
				DefaultBranch:   "main",
				IsProtected:     false,
				IsFeatureBranch: true,
			},
		},
		{
			name: "Push to Protected Branch from Config",
			env: map[string]string{
				"CI_COMMIT_REF_NAME": "dev",
			},
			config: `protected_branches: ["main", "dev", "release"]`,
			expected: Context{
				RefName:     "dev",
				IsProtected: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary config file
			if tt.config != "" {
				tmpfile, err := os.CreateTemp("", ".syac.yml")
				if err != nil {
					t.Fatal(err)
				}
				defer os.Remove(tmpfile.Name())
				if _, err := tmpfile.WriteString(tt.config); err != nil {
					t.Fatal(err)
				}
				if err := tmpfile.Close(); err != nil {
					t.Fatal(err)
				}
				// monkey patch the load config function to use our temp file
				wd, _ := os.Getwd()
				os.Chdir(tmpfile.Name())
				defer os.Chdir(wd)
			}

			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := LoadContext()

			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Compare only the fields we care about for simplicity
			if got.RefName != tt.expected.RefName {
				t.Errorf("expected RefName %q, got %q", tt.expected.RefName, got.RefName)
			}
			if got.IsProtected != tt.expected.IsProtected {
				t.Errorf("expected IsProtected %v, got %v", tt.expected.IsProtected, got.IsProtected)
			}
			if got.IsFeatureBranch != tt.expected.IsFeatureBranch {
				t.Errorf("expected IsFeatureBranch %v, got %v", tt.expected.IsFeatureBranch, got.IsFeatureBranch)
			}
		})
	}
}

func TestIsProtectedBranch(t *testing.T) {
	protected := []string{"main", "dev", "release", "staging"}
	tests := []struct {
		ref      string
		expected bool
	}{
		{"main", true},
		{"dev", true},
		{"release", true},
		{"staging", true},
		{"release/v1.0", true},
		{"feature/new-thing", false},
		{"bugfix/fix-bug", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.ref, func(t *testing.T) {
			got := isProtectedBranch(tt.ref, protected)
			if got != tt.expected {
				t.Errorf("isProtectedBranch(%q) = %v; want %v", tt.ref, got, tt.expected)
			}
		})
	}
}


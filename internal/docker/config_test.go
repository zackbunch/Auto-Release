package docker

import (
	"os"
	"testing"
)

func clearEnv(keys ...string) {
	for _, k := range keys {
		os.Unsetenv(k)
	}
}

func TestLoadConfigAndValidation(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantError   bool
		wantApp     string
		wantDocker  string
		wantEnv     string
		targetTag   string
		targetImage string
	}{
		{
			name: "valid config with overrides",
			env: map[string]string{
				"CI_REGISTRY_USER":      "zack",
				"CI_REGISTRY_PASSWORD":  "token",
				"CI_REGISTRY":           "registry.gitlab.com",
				"CI_REGISTRY_IMAGE":     "registry.gitlab.com/devops/syac",
				"CI_PROJECT_PATH":       "devops/syac/some-service",
				"CI_COMMIT_REF_NAME":    "gmarm-1234",
				"SYAC_DOCKERFILE":       "",
				"SYAC_APPLICATION_NAME": "custom-app",
				"SYAC_BUILD_EXTRA_ARGS": "--build-arg FOO=bar",
			},
			wantError:   false,
			wantApp:     "custom-app",
			wantDocker:  "Dockerfile",
			wantEnv:     "dev",
			targetTag:   "v1.2.3",
			targetImage: "registry.gitlab.com/devops/syac/dev/custom-app:v1.2.3",
		},
		{
			name: "fallback to image name from registry path",
			env: map[string]string{
				"CI_REGISTRY_USER":     "user",
				"CI_REGISTRY_PASSWORD": "pass",
				"CI_REGISTRY":          "registry.gitlab.com",
				"CI_REGISTRY_IMAGE":    "registry.gitlab.com/group/project",
				"CI_PROJECT_PATH":      "group/project",
				"CI_COMMIT_REF_NAME":   "main",
			},
			wantError:   false,
			wantApp:     "project",
			wantDocker:  "Dockerfile",
			wantEnv:     "prod",
			targetTag:   "latest",
			targetImage: "registry.gitlab.com/group/project/prod/project:latest",
		},
		{
			name: "missing required fields",
			env: map[string]string{
				"CI_REGISTRY": "registry.gitlab.com",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(
				"CI_REGISTRY_USER", "CI_REGISTRY_PASSWORD", "CI_REGISTRY",
				"CI_REGISTRY_IMAGE", "CI_PROJECT_PATH", "CI_COMMIT_REF_NAME",
				"SYAC_DOCKERFILE", "SYAC_APPLICATION_NAME", "SYAC_BUILD_EXTRA_ARGS",
			)
			for k, v := range tt.env {
				os.Setenv(k, v)
			}

			cfg, err := LoadConfig()
			if tt.wantError {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.ImageName() != tt.wantApp {
				t.Errorf("ImageName() = %s, want %s", cfg.ImageName(), tt.wantApp)
			}
			if cfg.Dockerfile != tt.wantDocker {
				t.Errorf("Dockerfile = %s, want %s", cfg.Dockerfile, tt.wantDocker)
			}
			if cfg.OpenShiftEnv != tt.wantEnv {
				t.Errorf("OpenShiftEnv = %s, want %s", cfg.OpenShiftEnv, tt.wantEnv)
			}
			if img := cfg.TargetImage(tt.targetTag); img != tt.targetImage {
				t.Errorf("TargetImage() = %s, want %s", img, tt.targetImage)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	input := "--build-arg A=1 --build-arg B=2"
	expected := []string{"--build-arg", "A=1", "--build-arg", "B=2"}
	result := parseArgs(input)

	if len(result) != len(expected) {
		t.Fatalf("expected %d args, got %d", len(expected), len(result))
	}
	for i := range expected {
		if result[i] != expected[i] {
			t.Errorf("arg %d = %s, want %s", i, result[i], expected[i])
		}
	}
}

func TestImageNameFallback(t *testing.T) {
	cfg := &Config{
		RegistryImagePath: "registry.gitlab.com/devops/syac/myapp",
	}
	if got := cfg.ImageName(); got != "myapp" {
		t.Errorf("ImageName fallback = %s, want myapp", got)
	}
}

func TestDeriveOpenShiftEnv(t *testing.T) {
	cases := map[string]string{
		"main":        "prod",
		"master":      "prod",
		"test":        "test",
		"feature/foo": "dev",
	}
	for ref, expected := range cases {
		if got := deriveOpenShiftEnv(ref); got != expected {
			t.Errorf("deriveOpenShiftEnv(%q) = %q, want %q", ref, got, expected)
		}
	}
}

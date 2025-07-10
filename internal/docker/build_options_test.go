package docker

import (
	"reflect"
	"syac/internal/ci"
	"testing"
)

func TestBuildOptionsFromContext(t *testing.T) {
	tests := []struct {
		name     string
		ctx      ci.Context
		env      map[string]string
		expected BuildOptions
	}{
		{
			name: "Feature Branch Build",
			ctx: ci.Context{
				RefName:       "feature/new-login",
				RegistryImage: "registry.example.com/my-group/my-app",
			},
			env: map[string]string{
				"CI_COMMIT_SHORT_SHA": "abcdefg",
			},
			expected: BuildOptions{
				Dockerfile:  "Dockerfile",
				ContextPath: ".",
				ImageName:   "my-app",
				TargetTag:   "abcdefg",
				FullImage:   "registry.example.com/my-group/my-app/dev/my-app:abcdefg",
				Push:        false, // dev does not push by default
				DryRun:      false,
			},
		},
		{
			name: "Feature Branch with Force Push",
			ctx: ci.Context{
				RefName:       "feature/new-login",
				RegistryImage: "registry.example.com/my-group/my-app",
			},
			env: map[string]string{
				"CI_COMMIT_SHORT_SHA": "abcdefg",
				"SYAC_FORCE_PUSH":     "true",
			},
			expected: BuildOptions{
				Dockerfile:  "Dockerfile",
				ContextPath: ".",
				ImageName:   "my-app",
				TargetTag:   "abcdefg",
				FullImage:   "registry.example.com/my-group/my-app/dev/my-app:abcdefg",
				Push:        true, // push is forced
				DryRun:      false,
			},
		},
		{
			name: "Merge Request to Dev (no version bump)",
			ctx: ci.Context{
				RefName:       "dev",
				IsMergeRequest: true,
				MRID:          "123",
				RegistryImage: "registry.example.com/my-group/my-app",
			},
			env: map[string]string{
				"CI_COMMIT_SHORT_SHA": "abcdefg",
			},
			expected: BuildOptions{
				Dockerfile:  "Dockerfile",
				ContextPath: ".",
				ImageName:   "my-app",
				TargetTag:   "abcdefg",
				FullImage:   "registry.example.com/my-group/my-app/dev/my-app:abcdefg",
				Push:        false,
			},
		},
		{
			name: "Push to Main Branch (Prod build)",
			ctx: ci.Context{
				RefName:       "main",
				RegistryImage: "registry.example.com/my-group/my-app",
			},
			env: map[string]string{
				"CI_COMMIT_SHORT_SHA": "1234567",
			},
			expected: BuildOptions{
				Dockerfile:  "Dockerfile",
				ContextPath: ".",
				ImageName:   "my-app",
				TargetTag:   "1234567",
				FullImage:   "registry.example.com/my-group/my-app/prod/my-app:1234567",
				Push:        true, // prod pushes by default
			},
		},
		{
			name: "Custom Dockerfile and Context",
			ctx: ci.Context{
				RefName:       "main",
				RegistryImage: "registry.example.com/my-group/my-app",
			},
			env: map[string]string{
				"CI_COMMIT_SHORT_SHA": "1234567",
				"SYAC_DOCKERFILE":     "./build/Dockerfile.prod",
				"SYAC_BUILD_CONTEXT":  "./server",
			},
			expected: BuildOptions{
				Dockerfile:  "./build/Dockerfile.prod",
				ContextPath: "./server",
				ImageName:   "my-app",
				TargetTag:   "1234567",
				FullImage:   "registry.example.com/my-group/my-app/prod/my-app:1234567",
				Push:        true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables for the current test
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			got, err := BuildOptionsFromContext(tt.ctx)
			if err != nil {
				t.Fatalf("BuildOptionsFromContext() error = %v", err)
			}

			// Zero out extra build args for comparison as it's not being tested here
			got.ExtraBuildArgs = nil

			if !reflect.DeepEqual(*got, tt.expected) {
				t.Errorf("BuildOptionsFromContext() = %+v, want %+v", *got, tt.expected)
			}
		})
	}
}

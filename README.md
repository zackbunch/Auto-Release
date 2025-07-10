# SYAC - Saved You A Click

SYAC is a GitLab-native build automation tool written in Go. It streamlines the container image build and promotion process for microservices deployed via GitLab CI/CD, targeting OpenShift environments.

## Purpose

The primary goal of SYAC is to automate the Docker image building and pushing process within a GitLab CI/CD pipeline. It intelligently determines image tags and whether to push an image based on the CI context (e.g., branch type, tags).

## Features

-   **Environment-driven Configuration:** Reads configuration from GitLab CI environment variables and supports local testing via `.env` files.
-   **Intelligent Image Tagging:**
    -   **Feature Branches:** Images are tagged with `CI_COMMIT_SHORT_SHA`.
    -   **Protected Branches (e.g., `main`, `release`):** Images are tagged with `CI_COMMIT_SHORT_SHA`.
    -   **`dev` Branch (Merge/Push):** Images are tagged as `rc.N` where `N` is provided via `SYAC_RC_NUMBER`.
    -   **Tag Pushes:** Images are tagged with the Git tag itself (e.g., `1.2.3`).
    -   **Latest Tag:** Pushes to the default branch (e.g., `main`) or new version tags will also be tagged as `latest`.
-   **Conditional Image Pushing:** Images are pushed to the registry based on the environment. By default, images for `dev` environment are not pushed unless explicitly forced (`SYAC_FORCE_PUSH=true`). Images for `prod`, `test`, `int` environments are always pushed.
-   **Dry Run Capability:** Supports a dry run mode (`SYAC_DRY_RUN=true`) that logs commands without executing them, useful for debugging and verification.
-   **Flexible Dockerfile and Build Context:** Allows specifying custom Dockerfile paths and build contexts via environment variables (`SYAC_DOCKERFILE`, `SYAC_BUILD_CONTEXT`).

## How It Works

SYAC analyzes the GitLab CI environment context to determine the appropriate build and push actions.

```mermaid
flowchart TD
    A[Start SYAC] --> B{Load CI Context};
    B --> C{Determine Build Options};
    C --> D{Is Tag Push?};
    D -- Yes --> E[Build and Push Image];
    D -- No --> F{Is Protected Branch?};
    F -- Yes --> G{Is 'dev' Branch (Merge/Push)?};
    G -- Yes --> H[Build and Push RC Image];
    G -- No --> I[Build and Push Image (Protected)];
    F -- No --> J{Is Feature Branch?};
    J -- Yes --> K[Build Image (Conditional Push)];
    J -- No --> L[Unknown Context - Skip];
    E --> M[End];
    H --> M;
    I --> M;
    K --> M;
    L --> M;
```

## Project Structure

```plaintext
syac/
├── main.go                     # Entry point for the CLI application.
├── Dockerfile                  # Dockerfile for building the SYAC application itself.
├── internal/
│   ├── ci/                     # Contains CI context loading and environment variable handling.
│   │   ├── context.go          # Loads and parses GitLab CI environment variables.
│   │   └── env.go              # Handles loading of .env files.
│   ├── config/                 # Application configuration.
│   │   └── config.go           # Loads SYAC specific configuration (e.g., protected branches).
│   ├── docker/                 # Core Docker operations.
│   │   ├── build.go            # Handles Docker image building.
│   │   ├── build_options.go    # Derives Docker build options from CI context and environment variables.
│   │   ├── cmd.go              # Utility for running shell commands (e.g., `docker`).
│   │   ├── push.go             # Handles Docker image pushing and registry login.
│   │   └── runner.go           # Main logic for orchestrating Docker build and push based on CI context.
│   └── version/                # Semantic versioning logic.
│       └── version.go          # Defines Version struct and methods for parsing and incrementing versions.
├── .env                        # Local override of GitLab CI variables for development.
├── go.mod                      # Go module definition.
├── go.sum                      # Go module checksums.
└── README.md                   # This documentation.
```

## Usage

To use SYAC, ensure your GitLab CI/CD pipeline sets the necessary environment variables (e.g., `CI_COMMIT_SHORT_SHA`, `CI_REGISTRY_IMAGE`).

For local testing, you can use a `.env` file and pass it via the `-env` flag:

```bash
go run main.go -env .env
```

To enable dry run mode:

```bash
SYAC_DRY_RUN=true go run main.go
```

To force a push even for `dev` environment:

```bash
SYAC_FORCE_PUSH=true go run main.go
```

## Environment Variables

| Variable              | Description                                                                 | Default Value |
| :-------------------- | :-------------------------------------------------------------------------- | :------------ |
| `SYAC_DOCKERFILE`     | Path to the Dockerfile to use for building the image.                       | `Dockerfile`  |
| `SYAC_BUILD_CONTEXT`  | Path to the build context for Docker.                                       | `.`           |
| `SYAC_BUILD_EXTRA_ARGS` | Additional arguments to pass to `docker build`. (e.g., `--no-cache`)      | (empty)       |
| `SYAC_APPLICATION_NAME` | The name of the application, used in the image tag.                         | Derived from `CI_REGISTRY_IMAGE` |
| `SYAC_FORCE_PUSH`     | Set to `true` to force image push even for `dev` environment.               | `false`       |
| `SYAC_DRY_RUN`        | Set to `true` to enable dry run mode (commands are logged but not executed).| `false`       |
| `SYAC_RC_NUMBER`      | The release candidate number (N) for `rc.N` tags on `dev` branch merges/pushes. | (required for RC) |
| `CI_COMMIT_SHORT_SHA` | GitLab CI variable: Short SHA of the current commit.                        | (required)    |
| `CI_REGISTRY_IMAGE`   | GitLab CI variable: Full path to the Docker image in the registry.          | (required)    |
| `CI_COMMIT_REF_NAME`  | GitLab CI variable: Name of the branch or tag.                              | (required)    |
| `CI_PIPELINE_SOURCE`  | GitLab CI variable: Source of the pipeline (e.g., `push`, `merge_request_event`). | (required)    |
| `CI_PROJECT_PATH`     | GitLab CI variable: Path to the project.                                    | (required)    |
| `CI_DEFAULT_BRANCH`   | GitLab CI variable: Default branch of the project.                          | (required)    |
| `CI_REGISTRY`         | GitLab CI variable: URL of the Docker registry.                             | (required for push) |
| `CI_REGISTRY_USER`    | GitLab CI variable: Username for Docker registry login.                     | (required for push) |
| `CI_REGISTRY_PASSWORD`| GitLab CI variable: Password for Docker registry login.                     | (required for push) |
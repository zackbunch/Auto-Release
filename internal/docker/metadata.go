package docker

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metadata struct {
	Image        string `json:"image"`
	Tag          string `json:"tag"`
	TargetImage  string `json:"target_image"`
	Project      string `json:"project"`
	Ref          string `json:"ref"`
	IsMergeToDev bool   `json:"is_merge_to_dev"`
	Sprint       string `json:"sprint,omitempty"`
	OpenShiftEnv string `json:"openshift_env"`
}

// WriteMetadata emits the syac_output.json metadata file for downstream pipeline jobs.
func WriteMetadata(cfg *Config) error {
	meta := Metadata{
		Image:        cfg.ImageName(),
		Tag:          cfg.Tag,
		TargetImage:  cfg.TargetImage(),
		Project:      cfg.Project,
		Ref:          cfg.Ref,
		IsMergeToDev: cfg.IsMergeToDev(),
		Sprint:       cfg.Sprint,
		OpenShiftEnv: cfg.OpenShiftEnv,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile("syac_output.json", data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

//go:build test

package docker

import (
	"fmt"
)

func BuildImage(cfg *Config) error {
	fmt.Println("[mock] BuildImage called")
	return nil
}

func PushImage(cfg *Config) error {
	fmt.Println("[mock] PushImage called")
	return nil
}

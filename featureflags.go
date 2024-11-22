package main

import (
	"os"
)

var FeatureFlagPlaceholder = false

func SetFeatureFlagPlaceholder(x bool) func() {
	oldValue := FeatureFlagPlaceholder
	FeatureFlagPlaceholder = x
	return func() { FeatureFlagPlaceholder = oldValue }
}

func init() {
	if os.Getenv("FEATURE_FLAG_PLACEHOLDER") == "1" {
		FeatureFlagPlaceholder = true
	}
}

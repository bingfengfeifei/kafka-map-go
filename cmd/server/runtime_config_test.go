package main

import "testing"

func TestBuildRuntimeConfigIncludesIframeModeWhenEnabled(t *testing.T) {
	config := buildRuntimeConfig(runtimeConfigOptions{
		iframeMode: "true",
	})

	if config["iframeMode"] != true {
		t.Fatalf("expected iframeMode=true, got %#v", config["iframeMode"])
	}
}

func TestBuildRuntimeConfigOmitsIframeModeWhenDisabled(t *testing.T) {
	config := buildRuntimeConfig(runtimeConfigOptions{
		iframeMode: "false",
	})

	if _, exists := config["iframeMode"]; exists {
		t.Fatalf("expected iframeMode to be omitted, got %#v", config["iframeMode"])
	}
}

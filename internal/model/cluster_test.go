package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestClusterJSONOmitsSaslPassword(t *testing.T) {
	cluster := Cluster{
		Name:         "local",
		Servers:      "kafka:9092",
		SaslUsername: "admin",
		SaslPassword: "secret",
	}

	data, err := json.Marshal(cluster)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if strings.Contains(string(data), "secret") || strings.Contains(string(data), "saslPassword") {
		t.Fatalf("cluster JSON leaked password: %s", string(data))
	}
}

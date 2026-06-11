package service

import (
	"testing"

	"github.com/bingfengfeifei/kafka-map-go/internal/model"
)

func TestNormalizeClusterDefaultsSecurityProtocol(t *testing.T) {
	cluster := &model.Cluster{
		Name:    " local ",
		Servers: " kafka:9092 ",
	}

	if err := normalizeCluster(cluster); err != nil {
		t.Fatalf("normalizeCluster() error = %v", err)
	}
	if cluster.Name != "local" {
		t.Fatalf("Name = %q, want local", cluster.Name)
	}
	if cluster.Servers != "kafka:9092" {
		t.Fatalf("Servers = %q, want kafka:9092", cluster.Servers)
	}
	if cluster.SecurityProtocol != "PLAINTEXT" {
		t.Fatalf("SecurityProtocol = %q, want PLAINTEXT", cluster.SecurityProtocol)
	}
}

func TestNormalizeClusterDefaultsSaslMechanism(t *testing.T) {
	cluster := &model.Cluster{
		Name:             "local",
		Servers:          "kafka:9092",
		SecurityProtocol: "sasl_plaintext",
		SaslUsername:     "admin",
		SaslPassword:     "secret",
	}

	if err := normalizeCluster(cluster); err != nil {
		t.Fatalf("normalizeCluster() error = %v", err)
	}
	if cluster.SecurityProtocol != "SASL_PLAINTEXT" {
		t.Fatalf("SecurityProtocol = %q, want SASL_PLAINTEXT", cluster.SecurityProtocol)
	}
	if cluster.SaslMechanism != "PLAIN" {
		t.Fatalf("SaslMechanism = %q, want PLAIN", cluster.SaslMechanism)
	}
}

func TestNormalizeClusterRejectsUnsupportedSaslMechanism(t *testing.T) {
	cluster := &model.Cluster{
		Name:             "local",
		Servers:          "kafka:9092",
		SecurityProtocol: "SASL_PLAINTEXT",
		SaslMechanism:    "GSSAPI",
		SaslUsername:     "admin",
		SaslPassword:     "secret",
	}

	if err := normalizeCluster(cluster); err == nil {
		t.Fatal("normalizeCluster() error = nil, want error")
	}
}

func TestNormalizeClusterRequiresSaslCredentials(t *testing.T) {
	cluster := &model.Cluster{
		Name:             "local",
		Servers:          "kafka:9092",
		SecurityProtocol: "SASL_SSL",
		SaslMechanism:    "SCRAM-SHA-256",
	}

	if err := normalizeCluster(cluster); err == nil {
		t.Fatal("normalizeCluster() error = nil, want error")
	}
}

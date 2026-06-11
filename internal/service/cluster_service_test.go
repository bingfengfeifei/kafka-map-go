package service

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/bingfengfeifei/kafka-map-go/internal/config"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
	"github.com/bingfengfeifei/kafka-map-go/pkg/database"
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

func TestBootstrapClustersCreatesClusterWithoutConnectionValidation(t *testing.T) {
	clusterRepo := newTestClusterRepository(t)
	clusterService := NewClusterService(clusterRepo, util.NewKafkaClientManager(), nil, nil, nil)

	err := clusterService.BootstrapClusters([]config.BootstrapClusterConfig{
		{
			Name:             " offline ",
			Servers:          "127.0.0.1:1",
			SecurityProtocol: "sasl_plaintext",
			AuthUsername:     " user ",
			AuthPassword:     "pass",
		},
	})
	if err != nil {
		t.Fatalf("bootstrap cluster: %v", err)
	}

	cluster, err := clusterRepo.FindByName("offline")
	if err != nil {
		t.Fatalf("find bootstrapped cluster: %v", err)
	}
	if cluster.Servers != "127.0.0.1:1" {
		t.Fatalf("expected normalized servers to be saved, got %q", cluster.Servers)
	}
	if cluster.SecurityProtocol != "SASL_PLAINTEXT" {
		t.Fatalf("expected security protocol to be normalized, got %q", cluster.SecurityProtocol)
	}
	if cluster.SaslMechanism != "PLAIN" {
		t.Fatalf("expected default sasl mechanism, got %q", cluster.SaslMechanism)
	}
	if cluster.SaslUsername != "user" || cluster.SaslPassword != "pass" {
		t.Fatalf("expected auth aliases to populate sasl credentials")
	}

	err = clusterService.BootstrapClusters([]config.BootstrapClusterConfig{
		{
			Name:    " offline ",
			Servers: "127.0.0.1:2",
		},
	})
	if err != nil {
		t.Fatalf("bootstrap duplicate cluster: %v", err)
	}

	clusters, err := clusterRepo.FindAll()
	if err != nil {
		t.Fatalf("find clusters: %v", err)
	}
	if len(clusters) != 1 {
		t.Fatalf("expected normalized duplicate name to be skipped, got %d clusters", len(clusters))
	}
}

func TestBootstrapClustersRejectsBlankBrokerList(t *testing.T) {
	clusterRepo := newTestClusterRepository(t)
	clusterService := NewClusterService(clusterRepo, util.NewKafkaClientManager(), nil, nil, nil)

	err := clusterService.BootstrapClusters([]config.BootstrapClusterConfig{
		{
			Name:    "bad",
			Servers: " , ",
		},
	})
	if err == nil {
		t.Fatal("expected blank broker list to be rejected")
	}
	if !strings.Contains(err.Error(), "broker server") {
		t.Fatalf("expected broker validation error, got %v", err)
	}

	clusters, err := clusterRepo.FindAll()
	if err != nil {
		t.Fatalf("find clusters: %v", err)
	}
	if len(clusters) != 0 {
		t.Fatalf("expected invalid bootstrap cluster not to be saved, got %d clusters", len(clusters))
	}
}

func newTestClusterRepository(t *testing.T) *repository.ClusterRepository {
	t.Helper()

	db, err := database.InitDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("init test database: %v", err)
	}

	return repository.NewClusterRepository(db)
}

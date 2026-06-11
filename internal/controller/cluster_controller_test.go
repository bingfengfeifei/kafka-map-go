package controller

import (
	"testing"

	"github.com/bingfengfeifei/kafka-map-go/internal/model"
)

func TestClusterRequestAppliesAuthAliases(t *testing.T) {
	name := " local "
	servers := " kafka:9092 "
	protocol := " sasl_plaintext "
	mechanism := " plain "
	username := " admin "
	password := "secret"

	req := clusterRequest{
		Name:             &name,
		Servers:          &servers,
		SecurityProtocol: &protocol,
		SaslMechanism:    &mechanism,
		AuthUsername:     &username,
		AuthPassword:     &password,
	}

	var cluster model.Cluster
	req.applyTo(&cluster)

	if cluster.Name != "local" {
		t.Fatalf("Name = %q, want local", cluster.Name)
	}
	if cluster.Servers != "kafka:9092" {
		t.Fatalf("Servers = %q, want kafka:9092", cluster.Servers)
	}
	if cluster.SecurityProtocol != "sasl_plaintext" {
		t.Fatalf("SecurityProtocol = %q, want sasl_plaintext", cluster.SecurityProtocol)
	}
	if cluster.SaslMechanism != "plain" {
		t.Fatalf("SaslMechanism = %q, want plain", cluster.SaslMechanism)
	}
	if cluster.SaslUsername != "admin" {
		t.Fatalf("SaslUsername = %q, want admin", cluster.SaslUsername)
	}
	if cluster.SaslPassword != "secret" {
		t.Fatalf("SaslPassword = %q, want secret", cluster.SaslPassword)
	}
}

func TestClusterRequestAppliesSaslFields(t *testing.T) {
	username := "sasl-user"
	password := "sasl-secret"
	req := clusterRequest{
		SaslUsername: &username,
		SaslPassword: &password,
	}

	var cluster model.Cluster
	req.applyTo(&cluster)

	if cluster.SaslUsername != "sasl-user" {
		t.Fatalf("SaslUsername = %q, want sasl-user", cluster.SaslUsername)
	}
	if cluster.SaslPassword != "sasl-secret" {
		t.Fatalf("SaslPassword = %q, want sasl-secret", cluster.SaslPassword)
	}
}

func TestParseClusterIDsAcceptsCommaSeparatedIDs(t *testing.T) {
	got, err := parseClusterIDs("1, 2,3")
	if err != nil {
		t.Fatalf("parseClusterIDs() error = %v", err)
	}

	want := []uint{1, 2, 3}
	if len(got) != len(want) {
		t.Fatalf("len(parseClusterIDs()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseClusterIDs()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestParseClusterIDsRejectsInvalidID(t *testing.T) {
	if _, err := parseClusterIDs("1,nope"); err == nil {
		t.Fatal("parseClusterIDs() error = nil, want error")
	}
}

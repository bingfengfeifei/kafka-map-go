package util

import (
	"crypto/tls"
	"fmt"
	"strings"
	"sync"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/xdg-go/scram"
)

type KafkaClientManager struct {
	adminClients map[uint]sarama.ClusterAdmin
	mu           sync.RWMutex
}

func NewKafkaClientManager() *KafkaClientManager {
	return &KafkaClientManager{
		adminClients: make(map[uint]sarama.ClusterAdmin),
	}
}

// GetAdminClient returns a cached admin client or creates a new one
func (m *KafkaClientManager) GetAdminClient(cluster *model.Cluster) (sarama.ClusterAdmin, error) {
	m.mu.RLock()
	client, exists := m.adminClients[cluster.ID]
	m.mu.RUnlock()

	if exists {
		return client, nil
	}

	// Create new admin client
	config := m.buildConfig(cluster)
	brokers := strings.Split(cluster.Servers, ",")

	admin, err := sarama.NewClusterAdmin(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	m.mu.Lock()
	m.adminClients[cluster.ID] = admin
	m.mu.Unlock()

	return admin, nil
}

// RemoveAdminClient removes and closes an admin client
func (m *KafkaClientManager) RemoveAdminClient(clusterID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.adminClients[clusterID]; exists {
		delete(m.adminClients, clusterID)
		return client.Close()
	}
	return nil
}

// CreateConsumer creates a new Kafka consumer
func (m *KafkaClientManager) CreateConsumer(cluster *model.Cluster) (sarama.Consumer, error) {
	config := m.buildConfig(cluster)
	brokers := strings.Split(cluster.Servers, ",")

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}
	return consumer, nil
}

// CreateClient creates a new Kafka client
func (m *KafkaClientManager) CreateClient(cluster *model.Cluster) (sarama.Client, error) {
	config := m.buildConfig(cluster)
	brokers := strings.Split(cluster.Servers, ",")

	client, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	return client, nil
}

// CreateProducer creates a new Kafka producer
func (m *KafkaClientManager) CreateProducer(cluster *model.Cluster) (sarama.SyncProducer, error) {
	config := m.buildConfig(cluster)
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	brokers := strings.Split(cluster.Servers, ",")

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	return producer, nil
}

// buildConfig builds Kafka configuration from cluster settings
func (m *KafkaClientManager) buildConfig(cluster *model.Cluster) *sarama.Config {
	config := sarama.NewConfig()
	config.Version = sarama.V3_4_0_0

	// Security protocol configuration
	switch cluster.SecurityProtocol {
	case "SASL_PLAINTEXT":
		config.Net.SASL.Enable = true
		m.configureSASL(config, cluster)
	case "SASL_SSL":
		config.Net.SASL.Enable = true
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true, // In production, should verify certificates
		}
		m.configureSASL(config, cluster)
	case "SSL":
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true,
		}
	case "PLAINTEXT":
		// No additional configuration needed
	}

	return config
}

// configureSASL configures SASL authentication
func (m *KafkaClientManager) configureSASL(config *sarama.Config, cluster *model.Cluster) {
	config.Net.SASL.User = cluster.SaslUsername
	config.Net.SASL.Password = cluster.SaslPassword

	switch cluster.SaslMechanism {
	case "PLAIN":
		config.Net.SASL.Mechanism = sarama.SASLTypePlaintext
	case "SCRAM-SHA-256":
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
		config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: scram.SHA256}
		}
	case "SCRAM-SHA-512":
		config.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
		config.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient {
			return &XDGSCRAMClient{HashGeneratorFcn: scram.SHA512}
		}
	}
}

// CloseAll closes all cached admin clients
func (m *KafkaClientManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, client := range m.adminClients {
		client.Close()
		delete(m.adminClients, id)
	}
}

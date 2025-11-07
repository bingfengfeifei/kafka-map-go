package service

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
)

type ClusterService struct {
	clusterRepo          *repository.ClusterRepository
	kafkaManager         *util.KafkaClientManager
	topicService         *TopicService
	brokerService        *BrokerService
	consumerGroupService *ConsumerGroupService
}

func NewClusterService(clusterRepo *repository.ClusterRepository, kafkaManager *util.KafkaClientManager, topicService *TopicService, brokerService *BrokerService, consumerGroupService *ConsumerGroupService) *ClusterService {
	return &ClusterService{
		clusterRepo:          clusterRepo,
		kafkaManager:         kafkaManager,
		topicService:         topicService,
		brokerService:        brokerService,
		consumerGroupService: consumerGroupService,
	}
}

// Create creates a new cluster
func (s *ClusterService) Create(cluster *model.Cluster) error {
	// Set default security protocol if not provided
	if cluster.SecurityProtocol == "" {
		cluster.SecurityProtocol = "PLAINTEXT"
	}

	// Validate connection
	if err := s.validateConnection(cluster); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}

	return s.clusterRepo.Create(cluster)
}

// Update updates an existing cluster
func (s *ClusterService) Update(cluster *model.Cluster) error {
	// Set default security protocol if not provided
	if cluster.SecurityProtocol == "" {
		cluster.SecurityProtocol = "PLAINTEXT"
	}

	// Validate connection
	if err := s.validateConnection(cluster); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}

	// Remove old admin client
	s.kafkaManager.RemoveAdminClient(cluster.ID)

	return s.clusterRepo.Update(cluster)
}

// Delete deletes a cluster
func (s *ClusterService) Delete(id uint) error {
	// Remove admin client
	s.kafkaManager.RemoveAdminClient(id)

	return s.clusterRepo.Delete(id)
}

// GetByID retrieves a cluster by ID
func (s *ClusterService) GetByID(id uint) (*model.Cluster, error) {
	return s.clusterRepo.FindByID(id)
}

// GetAll retrieves all clusters
func (s *ClusterService) GetAll() ([]model.Cluster, error) {
	return s.clusterRepo.FindAll()
}

// GetAllPaged retrieves clusters with pagination
func (s *ClusterService) GetAllPaged(pageIndex, pageSize int, name string) ([]model.Cluster, int64, error) {
	return s.clusterRepo.FindAllPaged(pageIndex, pageSize, name)
}

// GetClusterInfo retrieves cluster information with statistics
func (s *ClusterService) GetClusterInfo(id uint) (*dto.ClusterInfo, error) {
	cluster, err := s.clusterRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	info := &dto.ClusterInfo{
		ID:               cluster.ID,
		Name:             cluster.Name,
		Servers:          cluster.Servers,
		SecurityProtocol: cluster.SecurityProtocol,
		SaslMechanism:    cluster.SaslMechanism,
		SaslUsername:     cluster.SaslUsername,
		CreatedAt:        cluster.CreatedAt,
		UpdatedAt:        cluster.UpdatedAt,
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return info, nil
	}

	// Get brokers
	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return info, nil
	}
	info.BrokerCount = len(brokers)

	// Get topics
	topics, err := admin.ListTopics()
	if err != nil {
		return info, nil
	}
	info.TopicCount = len(topics)

	// Get consumer groups
	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return info, nil
	}
	info.ConsumerCount = len(groups)

	// Calculate statistics
	partitionCount := 0
	replicaCount := 0
	for _, topic := range topics {
		partitionCount += int(topic.NumPartitions)
		if topic.NumPartitions > 0 {
			replicaCount += int(topic.ReplicationFactor)
		}
	}

	info.PartitionCount = partitionCount
	info.ReplicaCount = replicaCount

	return info, nil
}

// validateConnection validates cluster connection
func (s *ClusterService) validateConnection(cluster *model.Cluster) error {
	brokers := strings.Split(cluster.Servers, ",")
	if len(brokers) == 0 {
		return fmt.Errorf("no brokers specified")
	}

	// Test network connectivity to first broker
	broker := strings.TrimSpace(brokers[0])
	conn, err := net.DialTimeout("tcp", broker, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to broker %s: %w", broker, err)
	}
	conn.Close()

	return nil
}

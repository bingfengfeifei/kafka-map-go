package service

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
)

type BrokerService struct {
	clusterRepo  *repository.ClusterRepository
	kafkaManager *util.KafkaClientManager
}

func NewBrokerService(clusterRepo *repository.ClusterRepository, kafkaManager *util.KafkaClientManager) *BrokerService {
	return &BrokerService{
		clusterRepo:  clusterRepo,
		kafkaManager: kafkaManager,
	}
}

// GetBrokers retrieves all brokers for a cluster
func (s *BrokerService) GetBrokers(clusterID uint) ([]dto.BrokerInfo, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster: %w", err)
	}

	var brokerInfos []dto.BrokerInfo
	for _, broker := range brokers {
		brokerInfos = append(brokerInfos, dto.BrokerInfo{
			ID:   broker.ID(),
			Host: broker.Addr(),
			Port: 0, // sarama doesn't expose port separately
			Rack: broker.Rack(),
		})
	}

	return brokerInfos, nil
}

// GetBrokerConfigs retrieves broker configurations
func (s *BrokerService) GetBrokerConfigs(clusterID uint, brokerID int32) ([]dto.BrokerConfig, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	// Describe broker configs
	configResource := sarama.ConfigResource{
		Type: sarama.BrokerResource,
		Name: fmt.Sprintf("%d", brokerID),
	}

	configs, err := admin.DescribeConfig(configResource)
	if err != nil {
		return nil, fmt.Errorf("failed to describe broker config: %w", err)
	}

	var brokerConfigs []dto.BrokerConfig
	for _, config := range configs {
		brokerConfigs = append(brokerConfigs, dto.BrokerConfig{
			Name:      config.Name,
			Value:     config.Value,
			ReadOnly:  config.ReadOnly,
			Sensitive: config.Sensitive,
		})
	}

	return brokerConfigs, nil
}

// UpdateBrokerConfigs updates broker configurations
func (s *BrokerService) UpdateBrokerConfigs(clusterID uint, brokerID int32, configs map[string]string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	// Build config entries map
	configMap := make(map[string]*string)
	for key, value := range configs {
		v := value
		configMap[key] = &v
	}

	// Alter broker configs
	err = admin.AlterConfig(sarama.BrokerResource, fmt.Sprintf("%d", brokerID), configMap, false)
	if err != nil {
		return fmt.Errorf("failed to alter broker config: %w", err)
	}

	return nil
}

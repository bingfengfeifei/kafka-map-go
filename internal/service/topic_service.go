package service

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
)

type TopicService struct {
	clusterRepo  *repository.ClusterRepository
	kafkaManager *util.KafkaClientManager
}

func NewTopicService(clusterRepo *repository.ClusterRepository, kafkaManager *util.KafkaClientManager) *TopicService {
	return &TopicService{
		clusterRepo:  clusterRepo,
		kafkaManager: kafkaManager,
	}
}

// GetTopics retrieves all topics for a cluster
func (s *TopicService) GetTopics(clusterID uint) ([]dto.TopicInfo, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	topics, err := admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	var topicInfos []dto.TopicInfo
	for name, detail := range topics {
		topicInfos = append(topicInfos, dto.TopicInfo{
			Name:            name,
			Partitions:      int(detail.NumPartitions),
			Replicas:        int(detail.ReplicationFactor),
			ISR:             int(detail.ReplicationFactor), // Simplified - actual ISR would need metadata
			UnderReplicated: false,                         // Simplified - would need partition metadata
		})
	}

	return topicInfos, nil
}

// GetTopicNames retrieves all topic names
func (s *TopicService) GetTopicNames(clusterID uint) ([]string, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	topics, err := admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	var names []string
	for name := range topics {
		names = append(names, name)
	}

	return names, nil
}

// GetTopicDetail retrieves detailed topic information
func (s *TopicService) GetTopicDetail(clusterID uint, topicName string) (*dto.TopicDetail, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	// Get topic metadata
	metadata, err := admin.DescribeTopics([]string{topicName})
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic: %w", err)
	}

	if len(metadata) == 0 {
		return nil, fmt.Errorf("topic not found: %s", topicName)
	}

	topicMetadata := metadata[0]

	// Get partition offsets
	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var partitions []dto.PartitionInfo
	for _, partition := range topicMetadata.Partitions {
		offset, err := client.GetOffset(topicName, partition.ID, sarama.OffsetNewest)
		if err != nil {
			offset = -1
		}

		partitions = append(partitions, dto.PartitionInfo{
			Partition: partition.ID,
			Leader:    partition.Leader,
			Replicas:  partition.Replicas,
			ISR:       partition.Isr,
			Offset:    offset,
		})
	}

	// Get topic configs
	configResource := sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topicName,
	}

	configs, err := admin.DescribeConfig(configResource)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic config: %w", err)
	}

	var topicConfigs []dto.TopicConfig
	for _, config := range configs {
		topicConfigs = append(topicConfigs, dto.TopicConfig{
			Name:      config.Name,
			Value:     config.Value,
			ReadOnly:  config.ReadOnly,
			Sensitive: config.Sensitive,
		})
	}

	return &dto.TopicDetail{
		Name:       topicName,
		Partitions: partitions,
		Configs:    topicConfigs,
	}, nil
}

// CreateTopic creates a new topic
func (s *TopicService) CreateTopic(clusterID uint, req *dto.CreateTopicRequest) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	// Convert config map to pointer map
	configEntries := make(map[string]*string)
	for key, value := range req.Configs {
		v := value
		configEntries[key] = &v
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     req.Partitions,
		ReplicationFactor: req.ReplicationFactor,
		ConfigEntries:     configEntries,
	}

	err = admin.CreateTopic(req.Name, topicDetail, false)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}

// DeleteTopics deletes multiple topics
func (s *TopicService) DeleteTopics(clusterID uint, topicNames []string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	err = admin.DeleteTopic(topicNames[0])
	if err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}

	// Delete remaining topics
	for i := 1; i < len(topicNames); i++ {
		admin.DeleteTopic(topicNames[i])
	}

	return nil
}

// ExpandPartitions expands the number of partitions for a topic
func (s *TopicService) ExpandPartitions(clusterID uint, topicName string, count int32) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	err = admin.CreatePartitions(topicName, count, nil, false)
	if err != nil {
		return fmt.Errorf("failed to expand partitions: %w", err)
	}

	return nil
}

// UpdateTopicConfigs updates topic configurations
func (s *TopicService) UpdateTopicConfigs(clusterID uint, topicName string, configs map[string]string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	// Build config map
	configMap := make(map[string]*string)
	for key, value := range configs {
		v := value
		configMap[key] = &v
	}

	err = admin.AlterConfig(sarama.TopicResource, topicName, configMap, false)
	if err != nil {
		return fmt.Errorf("failed to alter topic config: %w", err)
	}

	return nil
}

// GetMessages retrieves messages from a topic
func (s *TopicService) GetMessages(clusterID uint, topicName string, partition int32, offset int64, limit int) ([]dto.MessageRecord, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	consumer, err := s.kafkaManager.CreateConsumer(cluster)
	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topicName, partition, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to consume partition: %w", err)
	}
	defer partitionConsumer.Close()

	var messages []dto.MessageRecord
	timeout := time.After(5 * time.Second)

	for len(messages) < limit {
		select {
		case msg := <-partitionConsumer.Messages():
			headers := make(map[string]string)
			for _, header := range msg.Headers {
				headers[string(header.Key)] = string(header.Value)
			}

			messages = append(messages, dto.MessageRecord{
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       string(msg.Key),
				Value:     string(msg.Value),
				Timestamp: msg.Timestamp.UnixMilli(),
				Headers:   headers,
			})
		case <-timeout:
			return messages, nil
		}
	}

	return messages, nil
}

// SendMessage sends a message to a topic
func (s *TopicService) SendMessage(clusterID uint, topicName string, req *dto.SendMessageRequest) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	producer, err := s.kafkaManager.CreateProducer(cluster)
	if err != nil {
		return err
	}
	defer producer.Close()

	// Build message
	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Value: sarama.StringEncoder(req.Value),
	}

	if req.Key != "" {
		msg.Key = sarama.StringEncoder(req.Key)
	}

	// Add headers
	for key, value := range req.Headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	_, _, err = producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

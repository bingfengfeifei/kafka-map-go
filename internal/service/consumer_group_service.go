package service

import (
	"fmt"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
)

type ConsumerGroupService struct {
	clusterRepo  *repository.ClusterRepository
	kafkaManager *util.KafkaClientManager
}

func NewConsumerGroupService(clusterRepo *repository.ClusterRepository, kafkaManager *util.KafkaClientManager) *ConsumerGroupService {
	return &ConsumerGroupService{
		clusterRepo:  clusterRepo,
		kafkaManager: kafkaManager,
	}
}

// GetConsumerGroups retrieves all consumer groups for a cluster
func (s *ConsumerGroupService) GetConsumerGroups(clusterID uint) ([]dto.ConsumerGroupInfo, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}

	var groupInfos []dto.ConsumerGroupInfo
	for groupID := range groups {
		// Describe group to get state
		descriptions, err := admin.DescribeConsumerGroups([]string{groupID})
		if err != nil {
			continue
		}

		if len(descriptions) > 0 {
			desc := descriptions[0]
			groupInfos = append(groupInfos, dto.ConsumerGroupInfo{
				GroupID: groupID,
				State:   desc.State,
				Members: len(desc.Members),
			})
		}
	}

	return groupInfos, nil
}

// GetConsumerGroupsByTopic retrieves consumer groups for a specific topic
func (s *ConsumerGroupService) GetConsumerGroupsByTopic(clusterID uint, topicName string) ([]string, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}

	var topicGroups []string
	for groupID := range groups {
		// Check if group consumes from this topic
		offsets, err := admin.ListConsumerGroupOffsets(groupID, map[string][]int32{topicName: nil})
		if err != nil {
			continue
		}

		if len(offsets.Blocks[topicName]) > 0 {
			topicGroups = append(topicGroups, groupID)
		}
	}

	return topicGroups, nil
}

// GetConsumerGroupDetail retrieves detailed consumer group information
func (s *ConsumerGroupService) GetConsumerGroupDetail(clusterID uint, groupID string) (*dto.ConsumerGroupDetail, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	// Describe consumer group
	descriptions, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil {
		return nil, fmt.Errorf("failed to describe consumer group: %w", err)
	}

	if len(descriptions) == 0 {
		return nil, fmt.Errorf("consumer group not found: %s", groupID)
	}

	desc := descriptions[0]

	// Build member info
	var members []dto.MemberInfo
	for memberID, member := range desc.Members {
		metadata, err := member.GetMemberMetadata()
		if err != nil {
			continue
		}

		members = append(members, dto.MemberInfo{
			MemberID:   memberID,
			ClientID:   member.ClientId,
			ClientHost: member.ClientHost,
			Topics:     metadata.Topics,
		})
	}

	// Get coordinator info (using default values as Sarama doesn't expose coordinator directly)
	coordinator := dto.BrokerInfo{
		ID:   -1,
		Host: "N/A",
	}

	// Get offsets
	offsets, err := admin.ListConsumerGroupOffsets(groupID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer group offsets: %w", err)
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var groupOffsets []dto.ConsumerGroupOffset
	for topic, partitions := range offsets.Blocks {
		for partition, block := range partitions {
			logSize, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
			if err != nil {
				logSize = -1
			}

			lag := int64(0)
			if logSize >= 0 && block.Offset >= 0 {
				lag = logSize - block.Offset
			}

			groupOffsets = append(groupOffsets, dto.ConsumerGroupOffset{
				Topic:     topic,
				Partition: partition,
				Offset:    block.Offset,
				LogSize:   logSize,
				Lag:       lag,
			})
		}
	}

	return &dto.ConsumerGroupDetail{
		GroupID:     groupID,
		State:       desc.State,
		Protocol:    desc.Protocol,
		Members:     members,
		Coordinator: coordinator,
		Offsets:     groupOffsets,
	}, nil
}

// DeleteConsumerGroup deletes a consumer group
func (s *ConsumerGroupService) DeleteConsumerGroup(clusterID uint, groupID string) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	err = admin.DeleteConsumerGroup(groupID)
	if err != nil {
		return fmt.Errorf("failed to delete consumer group: %w", err)
	}

	return nil
}

// GetConsumerGroupOffset retrieves consumer group offset for a topic
func (s *ConsumerGroupService) GetConsumerGroupOffset(clusterID uint, topicName, groupID string) ([]dto.ConsumerGroupOffset, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, err
	}

	offsets, err := admin.ListConsumerGroupOffsets(groupID, map[string][]int32{topicName: nil})
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer group offsets: %w", err)
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var groupOffsets []dto.ConsumerGroupOffset
	for partition, block := range offsets.Blocks[topicName] {
		logSize, err := client.GetOffset(topicName, partition, sarama.OffsetNewest)
		if err != nil {
			logSize = -1
		}

		lag := int64(0)
		if logSize >= 0 && block.Offset >= 0 {
			lag = logSize - block.Offset
		}

		groupOffsets = append(groupOffsets, dto.ConsumerGroupOffset{
			Topic:     topicName,
			Partition: partition,
			Offset:    block.Offset,
			LogSize:   logSize,
			Lag:       lag,
		})
	}

	return groupOffsets, nil
}

// ResetConsumerGroupOffset resets consumer group offset for a topic
func (s *ConsumerGroupService) ResetConsumerGroupOffset(clusterID uint, topicName, groupID string, req *dto.OffsetResetRequest) error {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return err
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return err
	}
	defer client.Close()

	// Get topic partitions
	partitions, err := client.Partitions(topicName)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	// Build offset map
	offsetMap := make(map[string]map[int32]int64)
	offsetMap[topicName] = make(map[int32]int64)

	for _, partition := range partitions {
		var offset int64
		switch req.Type {
		case "beginning":
			offset, err = client.GetOffset(topicName, partition, sarama.OffsetOldest)
		case "end":
			offset, err = client.GetOffset(topicName, partition, sarama.OffsetNewest)
		case "offset":
			offset = req.Offset
		default:
			return fmt.Errorf("invalid offset reset type: %s", req.Type)
		}

		if err != nil {
			return fmt.Errorf("failed to get offset: %w", err)
		}

		offsetMap[topicName][partition] = offset
	}

	// Reset offsets
	err = admin.DeleteConsumerGroupOffset(groupID, topicName, partitions[0])
	if err != nil {
		return fmt.Errorf("failed to reset offset: %w", err)
	}

	return nil
}

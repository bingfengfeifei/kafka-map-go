package service

import (
	"fmt"
	"sort"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
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

func (s *ConsumerGroupService) GetConsumerGroups(clusterID uint) ([]dto.ConsumerGroupInfo, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	groupMap, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}
	if len(groupMap) == 0 {
		return []dto.ConsumerGroupInfo{}, nil
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	infos := make([]dto.ConsumerGroupInfo, 0, len(groupMap))
	for groupID := range groupMap {
		describe, err := admin.DescribeConsumerGroups([]string{groupID})
		if err != nil || len(describe) == 0 {
			continue
		}
		desc := describe[0]

		topicSet := make(map[string]struct{})
		for _, member := range desc.Members {
			meta, err := member.GetMemberMetadata()
			if err == nil && meta != nil {
				for _, topic := range meta.Topics {
					topicSet[topic] = struct{}{}
				}
			}
			assignment, err := member.GetMemberAssignment()
			if err == nil && assignment != nil {
				for topic := range assignment.Topics {
					topicSet[topic] = struct{}{}
				}
			}
		}

		offsets, err := admin.ListConsumerGroupOffsets(groupID, nil)
		var lagSum int64
		if err == nil && offsets != nil {
			for topic, partitions := range offsets.Blocks {
				topicSet[topic] = struct{}{}
				for partition, block := range partitions {
					if block == nil || block.Offset < 0 {
						continue
					}
					logSize, err := client.GetOffset(topic, partition, sarama.OffsetNewest)
					if err != nil || logSize < 0 {
						continue
					}
					lagSum += logSize - block.Offset
				}
			}
		}

		topics := make([]string, 0, len(topicSet))
		for topic := range topicSet {
			topics = append(topics, topic)
		}
		sort.Strings(topics)

		infos = append(infos, dto.ConsumerGroupInfo{
			GroupID: groupID,
			Topics:  topics,
			Lag:     lagSum,
		})
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].GroupID < infos[j].GroupID
	})

	return infos, nil
}

func (s *ConsumerGroupService) GetConsumerGroupsByTopic(clusterID uint, topicName string) ([]string, error) {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	groupMap, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}

	result := make([]string, 0)
	for groupID := range groupMap {
		offsets, err := admin.ListConsumerGroupOffsets(groupID, map[string][]int32{topicName: nil})
		if err != nil {
			continue
		}
		if partitions, ok := offsets.Blocks[topicName]; ok && len(partitions) > 0 {
			result = append(result, groupID)
		}
	}

	sort.Strings(result)
	return result, nil
}

func (s *ConsumerGroupService) CountConsumerGroups(clusterID uint) (int, error) {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return 0, err
	}

	groups, err := admin.ListConsumerGroups()
	if err != nil {
		return 0, fmt.Errorf("failed to list consumer groups: %w", err)
	}
	return len(groups), nil
}

func (s *ConsumerGroupService) GetConsumerGroupDetail(clusterID uint, groupID string) (*dto.ConsumerGroupDetail, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	describe, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil || len(describe) == 0 {
		return nil, fmt.Errorf("consumer group not found: %s", groupID)
	}
	desc := describe[0]

	var members []dto.MemberInfo
	for memberID, member := range desc.Members {
		meta, err := member.GetMemberMetadata()
		if err != nil || meta == nil {
			continue
		}
		members = append(members, dto.MemberInfo{
			MemberID:   memberID,
			ClientID:   member.ClientId,
			ClientHost: member.ClientHost,
			Topics:     meta.Topics,
		})
	}

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
			if block == nil {
				continue
			}
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

	sort.Slice(groupOffsets, func(i, j int) bool {
		if groupOffsets[i].Topic == groupOffsets[j].Topic {
			return groupOffsets[i].Partition < groupOffsets[j].Partition
		}
		return groupOffsets[i].Topic < groupOffsets[j].Topic
	})

	return &dto.ConsumerGroupDetail{
		GroupID:     groupID,
		State:       desc.State,
		Protocol:    desc.Protocol,
		Members:     members,
		Coordinator: dto.BrokerInfo{ID: -1, Host: "N/A"},
		Offsets:     groupOffsets,
	}, nil
}

func (s *ConsumerGroupService) DescribeConsumerGroup(clusterID uint, groupID string) ([]dto.ConsumerGroupDescribe, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	describe, err := admin.DescribeConsumerGroups([]string{groupID})
	if err != nil || len(describe) == 0 {
		return nil, fmt.Errorf("consumer group not found: %s", groupID)
	}
	desc := describe[0]

	offsets, err := admin.ListConsumerGroupOffsets(groupID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer group offsets: %w", err)
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var results []dto.ConsumerGroupDescribe
	for _, member := range desc.Members {
		assignment, err := member.GetMemberAssignment()
		if err != nil || assignment == nil {
			continue
		}
		for topic, partitions := range assignment.Topics {
			for _, partition := range partitions {
				describeEntry := buildDescribeEntry(client, offsets, groupID, topic, partition)
				describeEntry.ConsumerID = member.MemberId
				describeEntry.Host = member.ClientHost
				describeEntry.ClientID = member.ClientId
				results = append(results, describeEntry)
			}
		}
	}

	if len(results) == 0 && offsets != nil {
		for topic, partitions := range offsets.Blocks {
			for partition := range partitions {
				results = append(results, buildDescribeEntry(client, offsets, groupID, topic, partition))
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Topic == results[j].Topic {
			return results[i].Partition < results[j].Partition
		}
		return results[i].Topic < results[j].Topic
	})

	return results, nil
}

func (s *ConsumerGroupService) DeleteConsumerGroup(clusterID uint, groupID string) error {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return err
	}
	if err := admin.DeleteConsumerGroup(groupID); err != nil {
		return fmt.Errorf("failed to delete consumer group: %w", err)
	}
	return nil
}

func (s *ConsumerGroupService) GetConsumerGroupOffset(clusterID uint, topicName, groupID string) ([]dto.TopicOffset, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
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

	var result []dto.TopicOffset
	for partition, block := range offsets.Blocks[topicName] {
		entry := dto.TopicOffset{
			GroupID:   groupID,
			Topic:     topicName,
			Partition: partition,
		}
		if block != nil && block.Offset >= 0 {
			value := block.Offset
			entry.ConsumerOffset = &value
		}

		if beginning, err := client.GetOffset(topicName, partition, sarama.OffsetOldest); err == nil && beginning >= 0 {
			entry.BeginningOffset = &beginning
		}
		if end, err := client.GetOffset(topicName, partition, sarama.OffsetNewest); err == nil && end >= 0 {
			entry.EndOffset = &end
		}

		result = append(result, entry)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Partition < result[j].Partition
	})
	return result, nil
}

func (s *ConsumerGroupService) ResetConsumerGroupOffset(clusterID uint, topicName, groupID string, req *dto.OffsetResetRequest) error {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return err
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return err
	}
	defer client.Close()

	partitions, err := client.Partitions(topicName)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %w", err)
	}

	offsetManager, err := sarama.NewOffsetManagerFromClient(groupID, client)
	if err != nil {
		return fmt.Errorf("failed to create offset manager: %w", err)
	}
	defer offsetManager.Close()

	for _, partition := range partitions {
		var newOffset int64
		switch req.Type {
		case "beginning":
			newOffset, err = client.GetOffset(topicName, partition, sarama.OffsetOldest)
		case "end":
			newOffset, err = client.GetOffset(topicName, partition, sarama.OffsetNewest)
		case "offset":
			newOffset = req.Offset
		default:
			return fmt.Errorf("invalid offset reset type: %s", req.Type)
		}
		if err != nil {
			return fmt.Errorf("failed to fetch offset: %w", err)
		}

		if err := admin.DeleteConsumerGroupOffset(groupID, topicName, partition); err != nil {
			return fmt.Errorf("failed to delete consumer group offset: %w", err)
		}

		partitionManager, err := offsetManager.ManagePartition(topicName, partition)
		if err != nil {
			return fmt.Errorf("failed to manage partition offset: %w", err)
		}
		partitionManager.MarkOffset(newOffset, "")
		partitionManager.Close()
	}

	offsetManager.Commit()
	return nil
}

func (s *ConsumerGroupService) getClusterAndAdmin(clusterID uint) (*model.Cluster, sarama.ClusterAdmin, error) {
	cluster, err := s.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, nil, err
	}

	admin, err := s.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, nil, err
	}
	return cluster, admin, nil
}

func buildDescribeEntry(client sarama.Client, offsets *sarama.OffsetFetchResponse, groupID, topic string, partition int32) dto.ConsumerGroupDescribe {
	var currentOffset *int64
	if blocks, ok := offsets.Blocks[topic]; ok {
		if block, ok := blocks[partition]; ok && block != nil && block.Offset >= 0 {
			offsetValue := block.Offset
			currentOffset = &offsetValue
		}
	}

	var beginningPtr *int64
	if beginning, err := client.GetOffset(topic, partition, sarama.OffsetOldest); err == nil && beginning >= 0 {
		beginningPtr = &beginning
	}

	var endPtr *int64
	if end, err := client.GetOffset(topic, partition, sarama.OffsetNewest); err == nil && end >= 0 {
		endPtr = &end
	}

	var lagPtr *int64
	if currentOffset != nil && endPtr != nil {
		lag := *endPtr - *currentOffset
		lagPtr = &lag
	}

	return dto.ConsumerGroupDescribe{
		GroupID:       groupID,
		Topic:         topic,
		Partition:     partition,
		CurrentOffset: currentOffset,
		LogBeginning:  beginningPtr,
		LogEnd:        endPtr,
		Lag:           lagPtr,
	}
}

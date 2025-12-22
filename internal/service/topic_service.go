package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
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

// GetTopics retrieves all topics for a cluster with optional fuzzy filtering by name.
func (s *TopicService) GetTopics(clusterID uint, name string) ([]dto.TopicSummary, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	topicsDetail, err := admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	filter := strings.TrimSpace(name)
	filterLower := strings.ToLower(filter)

	var topicNames []string
	for topicName := range topicsDetail {
		if filter != "" && !strings.Contains(strings.ToLower(topicName), filterLower) {
			continue
		}
		topicNames = append(topicNames, topicName)
	}

	if len(topicNames) == 0 {
		return []dto.TopicSummary{}, nil
	}

	sort.Strings(topicNames)

	metadata, err := admin.DescribeTopics(topicNames)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topics: %w", err)
	}

	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster: %w", err)
	}

	brokerIDs := make([]int32, 0, len(brokers))
	for _, b := range brokers {
		brokerIDs = append(brokerIDs, b.ID())
	}

	topicSet := make(map[string]struct{}, len(topicNames))
	for _, n := range topicNames {
		topicSet[n] = struct{}{}
	}

	totalLogSize, _, err := describeLogDirs(admin, brokerIDs, topicSet)
	if err != nil {
		// fall back to -1 if unable to fetch
		totalLogSize = make(map[string]int64)
	}

	clusterIDStr := strconv.FormatUint(uint64(clusterID), 10)

	summaries := make([]dto.TopicSummary, 0, len(metadata))
	for _, md := range metadata {
		if md == nil || md.Err != sarama.ErrNoError {
			continue
		}
		replicaCount := 0
		if len(md.Partitions) > 0 && len(md.Partitions[0].Replicas) > 0 {
			replicaCount = len(md.Partitions[0].Replicas)
		}
		size := int64(-1)
		if v, ok := totalLogSize[md.Name]; ok {
			size = v
		}

		// Calculate total messages and last timestamp for each topic
		var totalMessages int64
		var lastTimestamp int64
		client, err := s.kafkaManager.CreateClient(cluster)
		if err != nil {
			continue
		}
		for _, partition := range md.Partitions {
			// Get the newest offset for this partition
			newestOffset, err := client.GetOffset(md.Name, partition.ID, sarama.OffsetNewest)
			if err != nil {
				continue
			}
			// Get the oldest offset for this partition
			oldestOffset, err := client.GetOffset(md.Name, partition.ID, sarama.OffsetOldest)
			if err != nil {
				continue
			}
			totalMessages += newestOffset - oldestOffset
			if newestOffset > oldestOffset {
				lastTimestamp = int64(time.Now().UnixMilli())
			}
		}
		client.Close()

		summaries = append(summaries, dto.TopicSummary{
			ClusterID:        clusterIDStr,
			Name:             md.Name,
			PartitionsCount:  len(md.Partitions),
			ReplicaCount:     replicaCount,
			TotalLogSize:     size,
			TotalMessages:    totalMessages,
			LastTimestamp:    lastTimestamp,
			ConsumerGroupCnt: 0, // can be enriched later; UI currently does not use this field
		})
	}

	return summaries, nil
}

// GetTopicNames retrieves topic names (optionally filtered).
func (s *TopicService) GetTopicNames(clusterID uint, name string) ([]string, error) {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	topics, err := admin.ListTopics()
	if err != nil {
		return nil, fmt.Errorf("failed to list topics: %w", err)
	}

	filter := strings.TrimSpace(name)
	filterLower := strings.ToLower(filter)

	var names []string
	for topicName := range topics {
		if filter != "" && !strings.Contains(strings.ToLower(topicName), filterLower) {
			continue
		}
		names = append(names, topicName)
	}
	sort.Strings(names)
	return names, nil
}

// CountTopics returns the number of topics in a cluster.
func (s *TopicService) CountTopics(clusterID uint) (int, error) {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return 0, err
	}

	topics, err := admin.ListTopics()
	if err != nil {
		return 0, fmt.Errorf("failed to list topics: %w", err)
	}
	return len(topics), nil
}

// GetTopicDetail retrieves topic metadata with partition offsets.
func (s *TopicService) GetTopicDetail(clusterID uint, topicName string) (*dto.TopicDetail, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	metadata, err := admin.DescribeTopics([]string{topicName})
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic: %w", err)
	}

	if len(metadata) == 0 || metadata[0] == nil || metadata[0].Err != sarama.ErrNoError {
		return nil, fmt.Errorf("topic not found: %s", topicName)
	}
	topicMetadata := metadata[0]

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	partitions := make([]dto.TopicPartitionSummary, 0, len(topicMetadata.Partitions))
	for _, partition := range topicMetadata.Partitions {
		if partition == nil {
			continue
		}
		beginning, err := client.GetOffset(topicName, partition.ID, sarama.OffsetOldest)
		if err != nil {
			beginning = -1
		}
		end, err := client.GetOffset(topicName, partition.ID, sarama.OffsetNewest)
		if err != nil {
			end = -1
		}
		partitions = append(partitions, dto.TopicPartitionSummary{
			Partition:       partition.ID,
			BeginningOffset: beginning,
			EndOffset:       end,
		})
	}
	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Partition < partitions[j].Partition
	})

	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster: %w", err)
	}

	brokerIDs := make([]int32, 0, len(brokers))
	for _, b := range brokers {
		brokerIDs = append(brokerIDs, b.ID())
	}

	topicSet := map[string]struct{}{topicName: {}}
	totalLogSize, _, err := describeLogDirs(admin, brokerIDs, topicSet)
	var size int64 = -1
	if err == nil {
		if v, ok := totalLogSize[topicName]; ok {
			size = v
		}
	}

	replicaCount := 0
	if len(topicMetadata.Partitions) > 0 && len(topicMetadata.Partitions[0].Replicas) > 0 {
		replicaCount = len(topicMetadata.Partitions[0].Replicas)
	}

	return &dto.TopicDetail{
		ClusterID:    strconv.FormatUint(uint64(clusterID), 10),
		Name:         topicName,
		ReplicaCount: replicaCount,
		TotalLogSize: size,
		Partitions:   partitions,
	}, nil
}

// GetTopicPartitions returns detailed partition information for a topic.
func (s *TopicService) GetTopicPartitions(clusterID uint, topicName string) ([]dto.TopicPartitionDetail, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	metadata, err := admin.DescribeTopics([]string{topicName})
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic: %w", err)
	}
	if len(metadata) == 0 || metadata[0] == nil || metadata[0].Err != sarama.ErrNoError {
		return nil, fmt.Errorf("topic not found: %s", topicName)
	}
	topicMetadata := metadata[0]

	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster: %w", err)
	}

	brokerLookup := make(map[int32]*sarama.Broker, len(brokers))
	brokerIDs := make([]int32, 0, len(brokers))
	for _, broker := range brokers {
		brokerLookup[broker.ID()] = broker
		brokerIDs = append(brokerIDs, broker.ID())
	}

	topicSet := map[string]struct{}{topicName: {}}
	_, partitionLogSize, err := describeLogDirs(admin, brokerIDs, topicSet)
	if err != nil {
		partitionLogSize = map[int32]map[string]map[int32]int64{}
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	partitions := make([]dto.TopicPartitionDetail, 0, len(topicMetadata.Partitions))
	for _, partition := range topicMetadata.Partitions {
		if partition == nil {
			continue
		}

		beginning, err := client.GetOffset(topicName, partition.ID, sarama.OffsetOldest)
		if err != nil {
			beginning = -1
		}
		end, err := client.GetOffset(topicName, partition.ID, sarama.OffsetNewest)
		if err != nil {
			end = -1
		}

		leaderNode := partition.Leader
		leader := dto.PartitionNode{ID: leaderNode, Host: "", Port: 0, LogSize: -1}
		if broker, ok := brokerLookup[leaderNode]; ok && broker != nil {
			host, port := splitHostPort(broker.Addr())
			leader.Host = host
			leader.Port = port
			if sizeMap, ok := partitionLogSize[leaderNode]; ok {
				if topicMap, ok := sizeMap[topicName]; ok {
					if size, ok := topicMap[partition.ID]; ok {
						leader.LogSize = size
					}
				}
			}
		}

		replicas := make([]dto.PartitionNode, 0, len(partition.Replicas))
		for _, replicaID := range partition.Replicas {
			node := dto.PartitionNode{ID: replicaID, Host: "", Port: 0, LogSize: -1}
			if broker, ok := brokerLookup[replicaID]; ok && broker != nil {
				host, port := splitHostPort(broker.Addr())
				node.Host = host
				node.Port = port
				if sizeMap, ok := partitionLogSize[replicaID]; ok {
					if topicMap, ok := sizeMap[topicName]; ok {
						if size, ok := topicMap[partition.ID]; ok {
							node.LogSize = size
						}
					}
				}
			}
			replicas = append(replicas, node)
		}

		isrNodes := make([]dto.PartitionNode, 0, len(partition.Isr))
		for _, isrID := range partition.Isr {
			node := dto.PartitionNode{ID: isrID, Host: "", Port: 0, LogSize: -1}
			if broker, ok := brokerLookup[isrID]; ok && broker != nil {
				host, port := splitHostPort(broker.Addr())
				node.Host = host
				node.Port = port
				if sizeMap, ok := partitionLogSize[isrID]; ok {
					if topicMap, ok := sizeMap[topicName]; ok {
						if size, ok := topicMap[partition.ID]; ok {
							node.LogSize = size
						}
					}
				}
			}
			isrNodes = append(isrNodes, node)
		}

		partitions = append(partitions, dto.TopicPartitionDetail{
			Partition:       partition.ID,
			Leader:          leader,
			Replicas:        replicas,
			ISR:             isrNodes,
			BeginningOffset: beginning,
			EndOffset:       end,
		})
	}

	sort.Slice(partitions, func(i, j int) bool {
		return partitions[i].Partition < partitions[j].Partition
	})

	return partitions, nil
}

// GetTopicBrokers returns broker level statistics for a given topic.
func (s *TopicService) GetTopicBrokers(clusterID uint, topicName string) ([]dto.BrokerDetail, error) {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	metadata, err := admin.DescribeTopics([]string{topicName})
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic: %w", err)
	}
	if len(metadata) == 0 || metadata[0] == nil || metadata[0].Err != sarama.ErrNoError {
		return nil, fmt.Errorf("topic not found: %s", topicName)
	}
	topicMetadata := metadata[0]

	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		return nil, fmt.Errorf("failed to describe cluster: %w", err)
	}

	brokerMap := make(map[int32]*dto.BrokerDetail, len(brokers))
	for _, broker := range brokers {
		host, port := splitHostPort(broker.Addr())
		brokerMap[broker.ID()] = &dto.BrokerDetail{
			ID:                 broker.ID(),
			Host:               host,
			Port:               port,
			LeaderPartitions:   []int32{},
			FollowerPartitions: []int32{},
		}
	}

	for _, partition := range topicMetadata.Partitions {
		if partition == nil {
			continue
		}
		if broker, ok := brokerMap[partition.Leader]; ok {
			broker.LeaderPartitions = append(broker.LeaderPartitions, partition.ID)
		}
		for _, replicaID := range partition.Replicas {
			if broker, ok := brokerMap[replicaID]; ok {
				broker.FollowerPartitions = append(broker.FollowerPartitions, partition.ID)
			}
		}
	}

	result := make([]dto.BrokerDetail, 0, len(brokerMap))
	for _, broker := range brokerMap {
		if len(broker.LeaderPartitions) == 0 && len(broker.FollowerPartitions) == 0 {
			continue
		}
		sort.Slice(broker.LeaderPartitions, func(i, j int) bool {
			return broker.LeaderPartitions[i] < broker.LeaderPartitions[j]
		})
		sort.Slice(broker.FollowerPartitions, func(i, j int) bool {
			return broker.FollowerPartitions[i] < broker.FollowerPartitions[j]
		})
		result = append(result, *broker)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})

	return result, nil
}

// GetTopicConsumerGroups returns consumer groups that are consuming the given topic.
func (s *TopicService) GetTopicConsumerGroups(clusterID uint, topicName string) ([]dto.TopicConsumerGroup, error) {
	cluster, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	groupMap, err := admin.ListConsumerGroups()
	if err != nil {
		return nil, fmt.Errorf("failed to list consumer groups: %w", err)
	}
	if len(groupMap) == 0 {
		return []dto.TopicConsumerGroup{}, nil
	}

	groupIDs := make([]string, 0, len(groupMap))
	for groupID := range groupMap {
		groupIDs = append(groupIDs, groupID)
	}

	client, err := s.kafkaManager.CreateClient(cluster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	type groupLag struct {
		groupID string
		lag     int64
	}

	consumerGroups := make([]groupLag, 0, len(groupIDs))
	for _, groupID := range groupIDs {
		offsets, err := admin.ListConsumerGroupOffsets(groupID, map[string][]int32{topicName: nil})
		if err != nil {
			continue
		}

		blocks, ok := offsets.Blocks[topicName]
		if !ok || len(blocks) == 0 {
			continue
		}

		var lagSum int64
		for partitionID, block := range blocks {
			if block == nil || block.Offset < 0 {
				continue
			}
			logSize, err := client.GetOffset(topicName, partitionID, sarama.OffsetNewest)
			if err != nil || logSize < 0 {
				continue
			}
			lagSum += logSize - block.Offset
		}

		consumerGroups = append(consumerGroups, groupLag{
			groupID: groupID,
			lag:     lagSum,
		})
	}

	sort.Slice(consumerGroups, func(i, j int) bool {
		return consumerGroups[i].groupID < consumerGroups[j].groupID
	})

	result := make([]dto.TopicConsumerGroup, 0, len(consumerGroups))
	for _, cg := range consumerGroups {
		result = append(result, dto.TopicConsumerGroup{
			GroupID: cg.groupID,
			Lag:     cg.lag,
		})
	}
	return result, nil
}

// GetTopicConfigs returns topic configuration entries.
func (s *TopicService) GetTopicConfigs(clusterID uint, topicName string) ([]dto.TopicConfig, error) {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return nil, err
	}

	configResource := sarama.ConfigResource{
		Type: sarama.TopicResource,
		Name: topicName,
	}

	configs, err := admin.DescribeConfig(configResource)
	if err != nil {
		return nil, fmt.Errorf("failed to describe topic config: %w", err)
	}

	result := make([]dto.TopicConfig, 0, len(configs))
	for _, entry := range configs {
		result = append(result, dto.TopicConfig{
			Name:      entry.Name,
			Value:     entry.Value,
			Default:   entry.Default,
			ReadOnly:  entry.ReadOnly,
			Sensitive: entry.Sensitive,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result, nil
}

// CreateTopic creates a new topic
func (s *TopicService) CreateTopic(clusterID uint, req *dto.CreateTopicRequest) error {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return err
	}

	configEntries := make(map[string]*string, len(req.Configs))
	for key, value := range req.Configs {
		v := value
		configEntries[key] = &v
	}

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     req.Partitions,
		ReplicationFactor: req.ReplicationFactor,
		ConfigEntries:     configEntries,
	}

	if err := admin.CreateTopic(req.Name, topicDetail, false); err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}
	return nil
}

// DeleteTopics deletes multiple topics
func (s *TopicService) DeleteTopics(clusterID uint, topicNames []string) error {
	if len(topicNames) == 0 {
		return errors.New("no topic specified")
	}

	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return err
	}

	if err := admin.DeleteTopic(topicNames[0]); err != nil {
		return fmt.Errorf("failed to delete topic: %w", err)
	}
	for i := 1; i < len(topicNames); i++ {
		_ = admin.DeleteTopic(topicNames[i])
	}
	return nil
}

// ExpandPartitions expands the number of partitions for a topic
func (s *TopicService) ExpandPartitions(clusterID uint, topicName string, count int32) error {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return err
	}

	if err := admin.CreatePartitions(topicName, count, nil, false); err != nil {
		return fmt.Errorf("failed to expand partitions: %w", err)
	}
	return nil
}

// UpdateTopicConfigs updates topic configurations
func (s *TopicService) UpdateTopicConfigs(clusterID uint, topicName string, configs map[string]string) error {
	_, admin, err := s.getClusterAndAdmin(clusterID)
	if err != nil {
		return err
	}

	configMap := make(map[string]*string, len(configs))
	for key, value := range configs {
		v := value
		configMap[key] = &v
	}

	if err := admin.AlterConfig(sarama.TopicResource, topicName, configMap, false); err != nil {
		return fmt.Errorf("failed to alter topic config: %w", err)
	}
	return nil
}

// findInJSON recursively searches for a key-value pair in a JSON structure.
// Returns true if the key exists anywhere in the structure with a value that contains the search string.
func findInJSON(data interface{}, key, value string) bool {
	switch v := data.(type) {
	case map[string]interface{}:
		for k, val := range v {
			if k == key {
				// Check if value contains the search string (fuzzy match)
				valStr := fmt.Sprintf("%v", val)
				if strings.Contains(valStr, value) {
					return true
				}
			}
			// Recursively search nested structures
			if findInJSON(val, key, value) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if findInJSON(item, key, value) {
				return true
			}
		}
	}
	return false
}

// GetMessages retrieves messages from a topic with optional filtering
func (s *TopicService) GetMessages(clusterID uint, topicName string, partition int32, offset int64, limit int, keyFilter, valueFilter, jsonKey, jsonValue string) ([]dto.MessageRecord, error) {
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

	// When filters are applied, we need to read more messages to find enough matches
	// Set a maximum number of messages to scan to avoid infinite loops
	maxScan := limit * 100
	if maxScan < 1000 {
		maxScan = 1000
	}
	scanned := 0

	// Check if JSON filtering is enabled (both key and value must be provided)
	jsonFilterEnabled := jsonKey != "" && jsonValue != ""

	var (
		messages    []dto.MessageRecord
		initialWait = 5 * time.Second
		idleWait    = 200 * time.Millisecond
		timer       = time.NewTimer(initialWait)
	)
	defer timer.Stop()

	resetTimer := func(d time.Duration) {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timer.Reset(d)
	}

	for len(messages) < limit && scanned < maxScan {
		select {
		case msg, ok := <-partitionConsumer.Messages():
			if !ok {
				return messages, nil
			}
			scanned++

			keyStr := string(msg.Key)
			valueStr := string(msg.Value)

			// Apply key filter (message key contains)
			if keyFilter != "" && !strings.Contains(keyStr, keyFilter) {
				resetTimer(idleWait)
				continue
			}

			// Apply value filter (message value contains)
			if valueFilter != "" && !strings.Contains(valueStr, valueFilter) {
				resetTimer(idleWait)
				continue
			}

			// Apply JSON field filter
			if jsonFilterEnabled {
				var jsonData interface{}
				if err := json.Unmarshal(msg.Value, &jsonData); err != nil {
					// Not valid JSON, skip this message
					resetTimer(idleWait)
					continue
				}
				if !findInJSON(jsonData, jsonKey, jsonValue) {
					resetTimer(idleWait)
					continue
				}
			}

			headers := make(map[string]string, len(msg.Headers))
			for _, header := range msg.Headers {
				headers[string(header.Key)] = string(header.Value)
			}
			messages = append(messages, dto.MessageRecord{
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       keyStr,
				Value:     valueStr,
				Timestamp: msg.Timestamp.UnixMilli(),
				Headers:   headers,
			})

			if len(messages) >= limit {
				return messages, nil
			}

			resetTimer(idleWait)
		case <-timer.C:
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

	msg := &sarama.ProducerMessage{
		Topic: topicName,
		Value: sarama.StringEncoder(req.Value),
	}

	if req.Key != "" {
		msg.Key = sarama.StringEncoder(req.Key)
	}

	for key, value := range req.Headers {
		msg.Headers = append(msg.Headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	if _, _, err := producer.SendMessage(msg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	return nil
}

func (s *TopicService) getClusterAndAdmin(clusterID uint) (*model.Cluster, sarama.ClusterAdmin, error) {
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

func describeLogDirs(admin sarama.ClusterAdmin, brokerIDs []int32, topicFilter map[string]struct{}) (map[string]int64, map[int32]map[string]map[int32]int64, error) {
	if len(brokerIDs) == 0 {
		return map[string]int64{}, map[int32]map[string]map[int32]int64{}, nil
	}

	logDirs, err := admin.DescribeLogDirs(brokerIDs)
	if err != nil {
		return nil, nil, err
	}

	total := make(map[string]int64)
	perBroker := make(map[int32]map[string]map[int32]int64)

	for brokerID, dirMetadata := range logDirs {
		for _, dir := range dirMetadata {
			for _, topic := range dir.Topics {
				if len(topicFilter) > 0 {
					if _, ok := topicFilter[topic.Topic]; !ok {
						continue
					}
				}
				if _, ok := total[topic.Topic]; !ok {
					total[topic.Topic] = 0
				}
				for _, partition := range topic.Partitions {
					total[topic.Topic] += partition.Size

					if _, ok := perBroker[brokerID]; !ok {
						perBroker[brokerID] = make(map[string]map[int32]int64)
					}
					if _, ok := perBroker[brokerID][topic.Topic]; !ok {
						perBroker[brokerID][topic.Topic] = make(map[int32]int64)
					}
					perBroker[brokerID][topic.Topic][partition.PartitionID] += partition.Size
				}
			}
		}
	}

	return total, perBroker, nil
}

func splitHostPort(addr string) (string, int32) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return addr, 0
	}
	port64, err := strconv.ParseInt(portStr, 10, 32)
	if err != nil {
		return host, 0
	}
	return host, int32(port64)
}

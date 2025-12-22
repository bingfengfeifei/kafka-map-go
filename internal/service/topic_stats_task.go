package service

import (
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
)

type TopicStatsTask struct {
	clusterRepo     *repository.ClusterRepository
	topicStatsRepo  *repository.TopicStatsRepository
	kafkaManager    *util.KafkaClientManager
	interval        time.Duration
	stopCh          chan struct{}
	refreshDoneCh   chan struct{}
	wg              sync.WaitGroup
}

func NewTopicStatsTask(clusterRepo *repository.ClusterRepository, topicStatsRepo *repository.TopicStatsRepository, kafkaManager *util.KafkaClientManager, interval time.Duration) *TopicStatsTask {
	return &TopicStatsTask{
		clusterRepo:    clusterRepo,
		topicStatsRepo: topicStatsRepo,
		kafkaManager:   kafkaManager,
		interval:       interval,
		stopCh:         make(chan struct{}),
		refreshDoneCh:  make(chan struct{}),
	}
}

func (t *TopicStatsTask) Start() {
	t.wg.Add(1)
	go t.run()
	log.Printf("[TopicStatsTask] Started, will refresh every %v", t.interval)
}

// WaitForFirstRefresh waits for the first refresh to complete
func (t *TopicStatsTask) WaitForFirstRefresh() {
	log.Println("[TopicStatsTask] Waiting for first refresh to complete...")
	<-t.refreshDoneCh
	log.Println("[TopicStatsTask] First refresh completed")
}

func (t *TopicStatsTask) Stop() {
	close(t.stopCh)
	t.wg.Wait()
	log.Println("[TopicStatsTask] Stopped")
}

// RefreshAllSync performs a synchronous refresh of all cluster stats
// This should be called on startup to populate the cache immediately
func (t *TopicStatsTask) RefreshAllSync() {
	log.Println("[TopicStatsTask] Running synchronous refresh...")
	t.refreshAll()
	log.Println("[TopicStatsTask] Synchronous refresh complete")
}

func (t *TopicStatsTask) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[TopicStatsTask] Recovered from panic: %v", r)
		}
		close(t.refreshDoneCh)
		t.wg.Done()
	}()

	// Run immediately on start
	log.Println("[TopicStatsTask] Running initial refresh...")
	t.refreshAll()
	log.Println("[TopicStatsTask] Initial refresh complete")
	close(t.refreshDoneCh) // Signal that first refresh is done

	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("[TopicStatsTask] Running scheduled refresh...")
			t.refreshAll()
		case <-t.stopCh:
			log.Println("[TopicStatsTask] Received stop signal")
			return
		}
	}
}

func (t *TopicStatsTask) refreshAll() {
	log.Println("[TopicStatsTask] Refreshing all cluster stats...")
	clusters, err := t.clusterRepo.FindAll()
	if err != nil {
		log.Printf("[TopicStatsTask] Failed to get clusters: %v", err)
		return
	}
	log.Printf("[TopicStatsTask] Found %d clusters", len(clusters))

	for _, cluster := range clusters {
		t.refreshCluster(cluster.ID)
	}
	log.Println("[TopicStatsTask] Refresh complete")
}

func (t *TopicStatsTask) refreshCluster(clusterID uint) {
	log.Printf("[TopicStatsTask] Refreshing cluster %d...", clusterID)
	_, admin, err := t.getClusterAndAdmin(clusterID)
	if err != nil {
		log.Printf("[TopicStatsTask] Failed to get admin for cluster %d: %v", clusterID, err)
		return
	}

	topicsDetail, err := admin.ListTopics()
	if err != nil {
		log.Printf("[TopicStatsTask] Failed to list topics for cluster %d: %v", clusterID, err)
		return
	}
	log.Printf("[TopicStatsTask] Found %d topics in cluster %d", len(topicsDetail), clusterID)

	// Get broker IDs for log size calculation
	brokers, _, err := admin.DescribeCluster()
	if err != nil {
		log.Printf("[TopicStatsTask] Failed to describe cluster %d: %v", clusterID, err)
		return
	}

	brokerIDs := make([]int32, 0, len(brokers))
	for _, b := range brokers {
		brokerIDs = append(brokerIDs, b.ID())
	}

	topicSet := make(map[string]struct{})
	for topicName := range topicsDetail {
		topicSet[topicName] = struct{}{}
	}

	// Create client for offset queries
	cluster, err := t.clusterRepo.FindByID(clusterID)
	if err != nil {
		return
	}

	client, err := t.kafkaManager.CreateClient(cluster)
	if err != nil {
		return
	}
	defer client.Close()

	consumer, err := t.kafkaManager.CreateConsumer(cluster)
	if err != nil {
		return
	}
	defer consumer.Close()

	now := time.Now()

	for topicName := range topicsDetail {
		// Get topic metadata
		metadata, err := admin.DescribeTopics([]string{topicName})
		if err != nil || len(metadata) == 0 {
			continue
		}
		md := metadata[0]
		if md.Err != sarama.ErrNoError {
			continue
		}

		// Calculate total messages
		var totalMessages int64
		var lastTimestamp int64

		for _, partition := range md.Partitions {
			newestOffset, err := client.GetOffset(topicName, partition.ID, sarama.OffsetNewest)
			if err != nil {
				continue
			}
			oldestOffset, err := client.GetOffset(topicName, partition.ID, sarama.OffsetOldest)
			if err != nil {
				continue
			}
			totalMessages += newestOffset - oldestOffset
		}

		// For lastTimestamp, we need to actually consume the last message
		// This is expensive, so we only do it if there are messages
		if totalMessages > 0 && len(md.Partitions) > 0 {
			lastTimestamp = t.getLastMessageTimestamp(client, consumer, topicName, md.Partitions)
			log.Printf("[TopicStatsTask] Topic %s: totalMessages=%d, lastTimestamp=%d (localtime=%s)",
				topicName, totalMessages, lastTimestamp, time.UnixMilli(lastTimestamp).Format("2006-01-02 15:04:05"))
		}

		// Save to database
		stats := &model.TopicStats{
			ClusterID:     clusterID,
			Name:          topicName,
			TotalMessages: totalMessages,
			LastTimestamp: lastTimestamp,
			UpdatedAt:     now,
		}

		if err := t.topicStatsRepo.Upsert(stats); err != nil {
			log.Printf("[TopicStatsTask] Failed to save stats for %s: %v", topicName, err)
		} else {
			log.Printf("[TopicStatsTask] Saved stats: %s, messages=%d, lastTimestamp=%d", topicName, totalMessages, lastTimestamp)
		}
	}
}

func (t *TopicStatsTask) getLastMessageTimestamp(client sarama.Client, consumer sarama.Consumer, topicName string, partitions []*sarama.PartitionMetadata) int64 {
	var maxTimestamp int64 = 0

	log.Printf("[TopicStatsTask] getLastMessageTimestamp: topic=%s, partitions=%d", topicName, len(partitions))

	for _, partition := range partitions {
		// Get the actual newest offset for this partition
		newestOffset, err := client.GetOffset(topicName, partition.ID, sarama.OffsetNewest)
		if err != nil || newestOffset <= 0 {
			log.Printf("[TopicStatsTask] Skipping partition %d: newestOffset=%d, err=%v", partition.ID, newestOffset, err)
			continue
		}

		log.Printf("[TopicStatsTask] Partition %d: newestOffset=%d", partition.ID, newestOffset)

		// Consume the last message (newestOffset - 1)
		// For offset N, messages are at 0, 1, 2, ..., N-1
		// So newestOffset - 1 gives us the last message
		partitionConsumer, err := consumer.ConsumePartition(topicName, partition.ID, newestOffset-1)
		if err != nil {
			log.Printf("[TopicStatsTask] Failed to consume partition %d: %v", partition.ID, err)
			continue
		}

		select {
		case msg := <-partitionConsumer.Messages():
			msgTimestamp := msg.Timestamp.UnixMilli()
			log.Printf("[TopicStatsTask] Got message: partition=%d, offset=%d, timestamp=%d (localtime=%s)",
				partition.ID, msg.Offset, msgTimestamp, time.UnixMilli(msgTimestamp).Format("2006-01-02 15:04:05"))
			if msgTimestamp > maxTimestamp {
				maxTimestamp = msgTimestamp
			}
		case <-time.After(2 * time.Second):
			log.Printf("[TopicStatsTask] Timeout waiting for message on partition %d", partition.ID)
			// Timeout, skip this partition
		}

		partitionConsumer.Close()
	}

	log.Printf("[TopicStatsTask] getLastMessageTimestamp result: %d (localtime=%s)",
		maxTimestamp, time.UnixMilli(maxTimestamp).Format("2006-01-02 15:04:05"))
	return maxTimestamp
}

// Helper to get cluster and admin (similar to TopicService)
func (t *TopicStatsTask) getClusterAndAdmin(clusterID uint) (*model.Cluster, sarama.ClusterAdmin, error) {
	cluster, err := t.clusterRepo.FindByID(clusterID)
	if err != nil {
		return nil, nil, err
	}

	admin, err := t.kafkaManager.GetAdminClient(cluster)
	if err != nil {
		return nil, nil, err
	}

	return cluster, admin, nil
}

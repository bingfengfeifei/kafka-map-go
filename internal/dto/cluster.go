package dto

import "time"

// ClusterInfo represents cluster information with statistics
type ClusterInfo struct {
	ID               uint      `json:"id"`
	Name             string    `json:"name"`
	Servers          string    `json:"servers"`
	SecurityProtocol string    `json:"securityProtocol"`
	SaslMechanism    string    `json:"saslMechanism"`
	SaslUsername     string    `json:"saslUsername"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	BrokerCount      int       `json:"brokerCount"`
	TopicCount       int       `json:"topicCount"`
	ConsumerCount    int       `json:"consumerCount"`
	PartitionCount   int       `json:"partitionCount"`
	ReplicaCount     int       `json:"replicaCount"`
}

// BrokerInfo is a lightweight broker descriptor used in various places.
type BrokerInfo struct {
	ID   int32  `json:"id"`
	Host string `json:"host"`
	Port int32  `json:"port"`
	Rack string `json:"rack,omitempty"`
}

// BrokerDetail represents broker information together with topic level statistics.
type BrokerDetail struct {
	ID                 int32   `json:"id"`
	Host               string  `json:"host"`
	Port               int32   `json:"port"`
	LeaderPartitions   []int32 `json:"leaderPartitions"`
	FollowerPartitions []int32 `json:"followerPartitions"`
}

// BrokerConfig represents broker configuration
type BrokerConfig struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Default   bool   `json:"_default"`
	ReadOnly  bool   `json:"readonly"`
	Sensitive bool   `json:"sensitive"`
}

// TopicSummary represents topic information shown in the topic list.
type TopicSummary struct {
	ClusterID        string `json:"clusterId"`
	Name             string `json:"name"`
	PartitionsCount  int    `json:"partitionsCount"`
	ReplicaCount     int    `json:"replicaCount"`
	TotalLogSize     int64  `json:"totalLogSize"`
	ConsumerGroupCnt int    `json:"consumerGroupCount"`
}

// TopicDetail represents topic overview with partition offsets.
type TopicDetail struct {
	ClusterID    string                  `json:"clusterId"`
	Name         string                  `json:"name"`
	ReplicaCount int                     `json:"replicaCount"`
	TotalLogSize int64                   `json:"totalLogSize"`
	Partitions   []TopicPartitionSummary `json:"partitions"`
}

// TopicPartitionSummary captures the offsets for a partition.
type TopicPartitionSummary struct {
	Partition       int32 `json:"partition"`
	BeginningOffset int64 `json:"beginningOffset"`
	EndOffset       int64 `json:"endOffset"`
}

// TopicPartitionDetail represents detailed partition information.
type TopicPartitionDetail struct {
	Partition       int32           `json:"partition"`
	Leader          PartitionNode   `json:"leader"`
	ISR             []PartitionNode `json:"isr"`
	Replicas        []PartitionNode `json:"replicas"`
	BeginningOffset int64           `json:"beginningOffset"`
	EndOffset       int64           `json:"endOffset"`
}

// PartitionNode represents metadata for a broker participating in a partition.
type PartitionNode struct {
	ID      int32  `json:"id"`
	Host    string `json:"host"`
	Port    int32  `json:"port"`
	LogSize int64  `json:"logSize"`
}

// TopicConfig represents topic configuration entries.
type TopicConfig struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Default   bool   `json:"_default"`
	ReadOnly  bool   `json:"readonly"`
	Sensitive bool   `json:"sensitive"`
}

// CreateTopicRequest represents topic creation request
type CreateTopicRequest struct {
	Name              string            `json:"name" binding:"required"`
	Partitions        int32             `json:"partitions" binding:"required,min=1"`
	ReplicationFactor int16             `json:"replicationFactor" binding:"required,min=1"`
	Configs           map[string]string `json:"configs"`
}

// TopicConsumerGroup represents aggregated lag information for a consumer group on a topic.
type TopicConsumerGroup struct {
	GroupID string `json:"groupId"`
	Lag     int64  `json:"lag"`
}

// TopicOffset represents offsets for a consumer group partition.
type TopicOffset struct {
	GroupID         string `json:"groupId"`
	Topic           string `json:"topic"`
	Partition       int32  `json:"partition"`
	ConsumerOffset  *int64 `json:"consumerOffset"`
	BeginningOffset *int64 `json:"beginningOffset"`
	EndOffset       *int64 `json:"endOffset"`
}

// ConsumerGroupDescribe represents detailed consumer group assignment information.
type ConsumerGroupDescribe struct {
	GroupID       string `json:"groupId"`
	Topic         string `json:"topic"`
	Partition     int32  `json:"partition"`
	CurrentOffset *int64 `json:"currentOffset"`
	LogBeginning  *int64 `json:"logBeginningOffset"`
	LogEnd        *int64 `json:"logEndOffset"`
	Lag           *int64 `json:"lag"`
	ConsumerID    string `json:"consumerId,omitempty"`
	Host          string `json:"host,omitempty"`
	ClientID      string `json:"clientId,omitempty"`
}

// ConsumerGroupInfo represents basic consumer group information for listing.
type ConsumerGroupInfo struct {
	GroupID string   `json:"groupId"`
	Topics  []string `json:"topics"`
	Lag     int64    `json:"lag"`
}

// ConsumerGroupDetail represents detailed consumer group information
type ConsumerGroupDetail struct {
	GroupID     string                `json:"groupId"`
	State       string                `json:"state"`
	Protocol    string                `json:"protocol"`
	Members     []MemberInfo          `json:"members"`
	Coordinator BrokerInfo            `json:"coordinator"`
	Offsets     []ConsumerGroupOffset `json:"offsets"`
}

// MemberInfo represents consumer group member information
type MemberInfo struct {
	MemberID   string   `json:"memberId"`
	ClientID   string   `json:"clientId"`
	ClientHost string   `json:"clientHost"`
	Topics     []string `json:"topics"`
}

// ConsumerGroupOffset represents consumer group offset information
type ConsumerGroupOffset struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	LogSize   int64  `json:"logSize"`
	Lag       int64  `json:"lag"`
}

// MessageRecord represents a Kafka message
type MessageRecord struct {
	Partition int32             `json:"partition"`
	Offset    int64             `json:"offset"`
	Key       string            `json:"key"`
	Value     string            `json:"value"`
	Timestamp int64             `json:"timestamp"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// SendMessageRequest represents message send request
type SendMessageRequest struct {
	Key     string            `json:"key"`
	Value   string            `json:"value" binding:"required"`
	Headers map[string]string `json:"headers"`
}

// OffsetResetRequest represents offset reset request
type OffsetResetRequest struct {
	Type   string `json:"type" binding:"required,oneof=beginning end offset"`
	Offset int64  `json:"offset"`
}

package dto

// ClusterInfo represents cluster information with statistics
type ClusterInfo struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Servers          string `json:"servers"`
	SecurityProtocol string `json:"securityProtocol"`
	SaslMechanism    string `json:"saslMechanism"`
	SaslUsername     string `json:"saslUsername"`
	BrokerCount      int    `json:"brokerCount"`
	TopicCount       int    `json:"topicCount"`
	PartitionCount   int    `json:"partitionCount"`
	ReplicaCount     int    `json:"replicaCount"`
}

// BrokerInfo represents broker information
type BrokerInfo struct {
	ID   int32  `json:"id"`
	Host string `json:"host"`
	Port int32  `json:"port"`
	Rack string `json:"rack,omitempty"`
}

// BrokerConfig represents broker configuration
type BrokerConfig struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	ReadOnly  bool   `json:"readOnly"`
	Sensitive bool   `json:"sensitive"`
}

// TopicInfo represents topic information
type TopicInfo struct {
	Name           string `json:"name"`
	Partitions     int    `json:"partitions"`
	Replicas       int    `json:"replicas"`
	ISR            int    `json:"isr"`
	UnderReplicated bool  `json:"underReplicated"`
}

// TopicDetail represents detailed topic information
type TopicDetail struct {
	Name       string            `json:"name"`
	Partitions []PartitionInfo   `json:"partitions"`
	Configs    []TopicConfig     `json:"configs"`
}

// PartitionInfo represents partition information
type PartitionInfo struct {
	Partition int32   `json:"partition"`
	Leader    int32   `json:"leader"`
	Replicas  []int32 `json:"replicas"`
	ISR       []int32 `json:"isr"`
	Offset    int64   `json:"offset"`
}

// TopicConfig represents topic configuration
type TopicConfig struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	ReadOnly  bool   `json:"readOnly"`
	Sensitive bool   `json:"sensitive"`
}

// CreateTopicRequest represents topic creation request
type CreateTopicRequest struct {
	Name              string            `json:"name" binding:"required"`
	Partitions        int32             `json:"partitions" binding:"required,min=1"`
	ReplicationFactor int16             `json:"replicationFactor" binding:"required,min=1"`
	Configs           map[string]string `json:"configs"`
}

// ConsumerGroupInfo represents consumer group information
type ConsumerGroupInfo struct {
	GroupID string `json:"groupId"`
	State   string `json:"state"`
	Members int    `json:"members"`
}

// ConsumerGroupDetail represents detailed consumer group information
type ConsumerGroupDetail struct {
	GroupID     string                  `json:"groupId"`
	State       string                  `json:"state"`
	Protocol    string                  `json:"protocol"`
	Members     []MemberInfo            `json:"members"`
	Coordinator BrokerInfo              `json:"coordinator"`
	Offsets     []ConsumerGroupOffset   `json:"offsets"`
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

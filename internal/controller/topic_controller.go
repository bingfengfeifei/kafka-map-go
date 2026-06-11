package controller

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/gin-gonic/gin"
)

type TopicController struct {
	topicService *service.TopicService
}

type messageQuery struct {
	partition   int32
	offset      int64
	limit       int
	keyFilter   string
	valueFilter string
	jsonKey     string
	jsonValue   string
}

func NewTopicController(topicService *service.TopicService) *TopicController {
	return &TopicController{topicService: topicService}
}

var errInvalidClusterID = errors.New("invalid cluster ID")

func parseClusterIDString(value string) (uint, error) {
	if value == "" {
		return 0, errInvalidClusterID
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func parseClusterIDValue(value any) (uint, error) {
	switch v := value.(type) {
	case nil:
		return 0, errInvalidClusterID
	case string:
		return parseClusterIDString(v)
	case float64:
		const maxUint32 = 1<<32 - 1
		if v < 0 || v > maxUint32 || v != float64(uint32(v)) {
			return 0, errInvalidClusterID
		}
		return uint(v), nil
	default:
		return 0, errInvalidClusterID
	}
}

func parseCreateTopicClusterID(ctx *gin.Context, req *dto.CreateTopicRequest) (uint, error) {
	if clusterID := ctx.Query("clusterId"); clusterID != "" {
		return parseClusterIDString(clusterID)
	}
	return parseClusterIDValue(req.ClusterID)
}

func normalizeCreateTopicRequest(req *dto.CreateTopicRequest) error {
	if req.Partitions == 0 {
		req.Partitions = req.NumPartitions
	}
	if req.Partitions < 1 {
		return errors.New("partitions must be at least 1")
	}
	return nil
}

func parseMessageQuery(ctx *gin.Context, defaultLimit int) messageQuery {
	partition, err := strconv.ParseInt(ctx.Query("partition"), 10, 32)
	if err != nil {
		partition = 0
	}

	offset, err := strconv.ParseInt(ctx.Query("offset"), 10, 64)
	if err != nil {
		offset = -1
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit <= 0 {
		limit, err = strconv.Atoi(ctx.Query("count"))
		if err != nil || limit <= 0 {
			limit = defaultLimit
		}
	}

	return messageQuery{
		partition:   int32(partition),
		offset:      offset,
		limit:       limit,
		keyFilter:   ctx.Query("keyFilter"),
		valueFilter: ctx.Query("valueFilter"),
		jsonKey:     ctx.Query("jsonKey"),
		jsonValue:   ctx.Query("jsonValue"),
	}
}

// GetTopics retrieves all topics for a cluster
func (c *TopicController) GetTopics(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	name := ctx.Query("name")

	topics, err := c.topicService.GetTopics(uint(clusterID), name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topics: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    topics,
	})
}

// GetTopicNames retrieves all topic names
func (c *TopicController) GetTopicNames(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	name := ctx.Query("name")

	names, err := c.topicService.GetTopicNames(uint(clusterID), name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topic names: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    names,
	})
}

// GetTopicDetail retrieves detailed topic information
func (c *TopicController) GetTopicDetail(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	detail, err := c.topicService.GetTopicDetail(uint(clusterID), topicName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topic detail: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    detail,
	})
}

// GetTopicPartitions returns partition level detail for a topic.
func (c *TopicController) GetTopicPartitions(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	partitions, err := c.topicService.GetTopicPartitions(uint(clusterID), topicName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topic partitions: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    partitions,
	})
}

// GetTopicBrokers returns broker level statistics for a topic.
func (c *TopicController) GetTopicBrokers(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	brokers, err := c.topicService.GetTopicBrokers(uint(clusterID), topicName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topic brokers: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    brokers,
	})
}

// GetTopicConsumerGroups returns consumer groups lag information for a topic.
func (c *TopicController) GetTopicConsumerGroups(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	groups, err := c.topicService.GetTopicConsumerGroups(uint(clusterID), topicName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topic consumer groups: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    groups,
	})
}

// CreateTopic creates a new topic
func (c *TopicController) CreateTopic(ctx *gin.Context) {
	var req dto.CreateTopicRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := normalizeCreateTopicRequest(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	clusterID, err := parseCreateTopicClusterID(ctx, &req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	if err := c.topicService.CreateTopic(clusterID, &req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed to create topic: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Topic created successfully",
	})
}

// DeleteTopics deletes multiple topics
func (c *TopicController) DeleteTopics(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	var topicNames []string
	if err := ctx.ShouldBindJSON(&topicNames); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.topicService.DeleteTopics(uint(clusterID), topicNames); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete topics: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Topics deleted successfully",
	})
}

// ExpandPartitions expands topic partitions
func (c *TopicController) ExpandPartitions(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")

	var req struct {
		Count int32 `json:"count" binding:"required,min=1"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.topicService.ExpandPartitions(uint(clusterID), topicName, req.Count); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to expand partitions: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Partitions expanded successfully",
	})
}

// GetTopicConfigs returns topic configuration entries.
func (c *TopicController) GetTopicConfigs(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	configs, err := c.topicService.GetTopicConfigs(uint(clusterID), topicName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve topic configs: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    configs,
	})
}

// UpdateTopicConfigs updates topic configurations
func (c *TopicController) UpdateTopicConfigs(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")

	var configs map[string]string
	if err := ctx.ShouldBindJSON(&configs); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.topicService.UpdateTopicConfigs(uint(clusterID), topicName, configs); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update topic configs: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Topic configs updated successfully",
	})
}

// GetMessages retrieves messages from a topic
func (c *TopicController) GetMessages(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	query := parseMessageQuery(ctx, 100)

	messages, err := c.topicService.GetMessages(uint(clusterID), topicName, query.partition, query.offset, query.limit, query.keyFilter, query.valueFilter, query.jsonKey, query.jsonValue)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve messages: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    messages,
	})
}

// GetMessagesLive streams newly consumed messages from a topic partition.
func (c *TopicController) GetMessagesLive(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	query := parseMessageQuery(ctx, 20)

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no")

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Streaming unsupported",
		})
		return
	}

	sendEvent := func(name string, data any) {
		ctx.SSEvent(name, data)
		flusher.Flush()
	}

	currentOffset := query.offset
	sendEvent("connected", gin.H{"status": "connected"})

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		default:
		}

		messages, err := c.topicService.GetMessages(uint(clusterID), topicName, query.partition, currentOffset, query.limit, query.keyFilter, query.valueFilter, query.jsonKey, query.jsonValue)
		if err != nil {
			sendEvent("error", gin.H{"message": "Failed to retrieve messages: " + err.Error()})
			return
		}

		beginningOffset, endOffset, err := c.topicService.GetPartitionOffsets(uint(clusterID), topicName, query.partition)
		if err != nil {
			beginningOffset = -1
			endOffset = -1
		}

		if len(messages) == 0 {
			sendEvent("ping", gin.H{
				"partition":       query.partition,
				"beginningOffset": beginningOffset,
				"endOffset":       endOffset,
				"timestamp":       time.Now().UnixMilli(),
			})
			continue
		}

		maxOffset := currentOffset
		for _, message := range messages {
			if message.Offset >= maxOffset {
				maxOffset = message.Offset + 1
			}
		}
		currentOffset = maxOffset

		sendEvent("topic-message-event", gin.H{
			"partition":       query.partition,
			"beginningOffset": beginningOffset,
			"endOffset":       endOffset,
			"messages":        messages,
		})
	}
}

// SendMessage sends a message to a topic
func (c *TopicController) SendMessage(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")

	var req dto.SendMessageRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.topicService.SendMessage(uint(clusterID), topicName, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to send message: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Message sent successfully",
	})
}

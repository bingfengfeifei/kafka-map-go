package controller

import (
	"net/http"
	"strconv"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/gin-gonic/gin"
)

type TopicController struct {
	topicService *service.TopicService
}

func NewTopicController(topicService *service.TopicService) *TopicController {
	return &TopicController{topicService: topicService}
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
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	var req dto.CreateTopicRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.topicService.CreateTopic(uint(clusterID), &req); err != nil {
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

	partition, err := strconv.ParseInt(ctx.Query("partition"), 10, 32)
	if err != nil {
		partition = 0
	}

	offset, err := strconv.ParseInt(ctx.Query("offset"), 10, 64)
	if err != nil {
		offset = -1 // Latest
	}

	limit, err := strconv.Atoi(ctx.Query("limit"))
	if err != nil || limit <= 0 {
		limit = 100
	}

	messages, err := c.topicService.GetMessages(uint(clusterID), topicName, int32(partition), offset, limit)
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

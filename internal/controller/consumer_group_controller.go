package controller

import (
	"net/http"
	"strconv"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/gin-gonic/gin"
)

type ConsumerGroupController struct {
	consumerGroupService *service.ConsumerGroupService
}

func NewConsumerGroupController(consumerGroupService *service.ConsumerGroupService) *ConsumerGroupController {
	return &ConsumerGroupController{consumerGroupService: consumerGroupService}
}

// GetConsumerGroups retrieves all consumer groups for a cluster
func (c *ConsumerGroupController) GetConsumerGroups(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	groups, err := c.consumerGroupService.GetConsumerGroups(uint(clusterID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve consumer groups: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    groups,
	})
}

// GetConsumerGroupsByTopic retrieves consumer groups for a specific topic
func (c *ConsumerGroupController) GetConsumerGroupsByTopic(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")

	groups, err := c.consumerGroupService.GetConsumerGroupsByTopic(uint(clusterID), topicName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve consumer groups: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    groups,
	})
}

// GetConsumerGroupDetail retrieves detailed consumer group information
func (c *ConsumerGroupController) GetConsumerGroupDetail(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	groupID := ctx.Param("groupId")

	detail, err := c.consumerGroupService.GetConsumerGroupDetail(uint(clusterID), groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve consumer group detail: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    detail,
	})
}

// DescribeConsumerGroup returns partition level assignment information for a consumer group.
func (c *ConsumerGroupController) DescribeConsumerGroup(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	groupID := ctx.Param("groupId")

	describe, err := c.consumerGroupService.DescribeConsumerGroup(uint(clusterID), groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to describe consumer group: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    describe,
	})
}

// DeleteConsumerGroup deletes a consumer group
func (c *ConsumerGroupController) DeleteConsumerGroup(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	groupID := ctx.Param("groupId")

	if err := c.consumerGroupService.DeleteConsumerGroup(uint(clusterID), groupID); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete consumer group: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Consumer group deleted successfully",
	})
}

// GetConsumerGroupOffset retrieves consumer group offset for a topic
func (c *ConsumerGroupController) GetConsumerGroupOffset(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	groupID := ctx.Param("groupId")

	offsets, err := c.consumerGroupService.GetConsumerGroupOffset(uint(clusterID), topicName, groupID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve consumer group offset: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    offsets,
	})
}

// ResetConsumerGroupOffset resets consumer group offset for a topic
func (c *ConsumerGroupController) ResetConsumerGroupOffset(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	topicName := ctx.Param("topic")
	groupID := ctx.Param("groupId")

	var req dto.OffsetResetRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.consumerGroupService.ResetConsumerGroupOffset(uint(clusterID), topicName, groupID, &req); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to reset consumer group offset: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Consumer group offset reset successfully",
	})
}

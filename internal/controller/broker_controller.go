package controller

import (
	"net/http"
	"strconv"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/gin-gonic/gin"
)

type BrokerController struct {
	brokerService *service.BrokerService
}

func NewBrokerController(brokerService *service.BrokerService) *BrokerController {
	return &BrokerController{brokerService: brokerService}
}

// GetBrokers retrieves all brokers for a cluster
func (c *BrokerController) GetBrokers(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	brokers, err := c.brokerService.GetBrokers(uint(clusterID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve brokers: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    brokers,
	})
}

// GetBrokerConfigs retrieves broker configurations
func (c *BrokerController) GetBrokerConfigs(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	brokerID, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid broker ID",
		})
		return
	}

	configs, err := c.brokerService.GetBrokerConfigs(uint(clusterID), int32(brokerID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve broker configs: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    configs,
	})
}

// UpdateBrokerConfigs updates broker configurations
func (c *BrokerController) UpdateBrokerConfigs(ctx *gin.Context) {
	clusterID, err := strconv.ParseUint(ctx.Query("clusterId"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	brokerID, err := strconv.ParseInt(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid broker ID",
		})
		return
	}

	var configs map[string]string
	if err := ctx.ShouldBindJSON(&configs); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.brokerService.UpdateBrokerConfigs(uint(clusterID), int32(brokerID), configs); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update broker configs: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Broker configs updated successfully",
	})
}

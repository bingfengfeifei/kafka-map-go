package controller

import (
	"net/http"
	"strconv"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/gin-gonic/gin"
)

type ClusterController struct {
	clusterService *service.ClusterService
}

func NewClusterController(clusterService *service.ClusterService) *ClusterController {
	return &ClusterController{clusterService: clusterService}
}

// GetClusters retrieves all clusters
func (c *ClusterController) GetClusters(ctx *gin.Context) {
	clusters, err := c.clusterService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to retrieve clusters: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    clusters,
	})
}

// GetCluster retrieves a cluster by ID
func (c *ClusterController) GetCluster(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	info, err := c.clusterService.GetClusterInfo(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.Response{
			Code:    http.StatusNotFound,
			Message: "Cluster not found: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    info,
	})
}

// CreateCluster creates a new cluster
func (c *ClusterController) CreateCluster(ctx *gin.Context) {
	var cluster model.Cluster
	if err := ctx.ShouldBindJSON(&cluster); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if err := c.clusterService.Create(&cluster); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed to create cluster: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Cluster created successfully",
		Data:    cluster,
	})
}

// UpdateCluster updates an existing cluster
func (c *ClusterController) UpdateCluster(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	var cluster model.Cluster
	if err := ctx.ShouldBindJSON(&cluster); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	cluster.ID = uint(id)
	if err := c.clusterService.Update(&cluster); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Failed to update cluster: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Cluster updated successfully",
		Data:    cluster,
	})
}

// DeleteCluster deletes a cluster
func (c *ClusterController) DeleteCluster(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID",
		})
		return
	}

	if err := c.clusterService.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, dto.Response{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete cluster: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Cluster deleted successfully",
	})
}

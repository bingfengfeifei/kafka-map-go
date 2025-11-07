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

type updateClusterRequest struct {
	Name             *string `json:"name"`
	Servers          *string `json:"servers"`
	SecurityProtocol *string `json:"securityProtocol"`
	SaslMechanism    *string `json:"saslMechanism"`
	SaslUsername     *string `json:"saslUsername"`
	SaslPassword     *string `json:"saslPassword"`
	AuthUsername     *string `json:"authUsername"`
	AuthPassword     *string `json:"authPassword"`
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

// GetClustersPaged retrieves clusters with pagination
func (c *ClusterController) GetClustersPaged(ctx *gin.Context) {
	pageIndex, err := strconv.Atoi(ctx.DefaultQuery("pageIndex", "1"))
	if err != nil || pageIndex < 1 {
		pageIndex = 1
	}

	pageSize, err := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	name := ctx.Query("name")

	clusters, total, err := c.clusterService.GetAllPaged(pageIndex, pageSize, name)
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
		Data: map[string]interface{}{
			"items":     clusters,
			"total":     total,
			"pageIndex": pageIndex,
			"pageSize":  pageSize,
		},
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

	var req updateClusterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if req.Name == nil && req.Servers == nil && req.SecurityProtocol == nil &&
		req.SaslMechanism == nil && req.SaslUsername == nil && req.SaslPassword == nil &&
		req.AuthUsername == nil && req.AuthPassword == nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "No fields to update",
		})
		return
	}

	cluster, err := c.clusterService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, dto.Response{
			Code:    http.StatusNotFound,
			Message: "Cluster not found: " + err.Error(),
		})
		return
	}

	if req.Name != nil {
		cluster.Name = *req.Name
	}
	if req.Servers != nil {
		cluster.Servers = *req.Servers
	}
	if req.SecurityProtocol != nil {
		cluster.SecurityProtocol = *req.SecurityProtocol
	}
	if req.SaslMechanism != nil {
		cluster.SaslMechanism = *req.SaslMechanism
	}
	if req.SaslUsername != nil {
		cluster.SaslUsername = *req.SaslUsername
	}
	if req.SaslPassword != nil {
		cluster.SaslPassword = *req.SaslPassword
	}
	if req.AuthUsername != nil {
		cluster.SaslUsername = *req.AuthUsername
	}
	if req.AuthPassword != nil {
		cluster.SaslPassword = *req.AuthPassword
	}

	if err := c.clusterService.Update(cluster); err != nil {
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

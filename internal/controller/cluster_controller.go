package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bingfengfeifei/kafka-map-go/internal/dto"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/gin-gonic/gin"
)

type ClusterController struct {
	clusterService *service.ClusterService
}

type clusterRequest struct {
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
	var req clusterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	cluster := model.Cluster{}
	req.applyTo(&cluster)
	if strings.TrimSpace(cluster.Name) == "" || strings.TrimSpace(cluster.Servers) == "" {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Cluster name and servers are required",
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

	var req clusterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	if !req.hasChanges() {
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

	req.applyTo(cluster)

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

func (r clusterRequest) hasChanges() bool {
	return r.Name != nil || r.Servers != nil || r.SecurityProtocol != nil ||
		r.SaslMechanism != nil || r.SaslUsername != nil || r.SaslPassword != nil ||
		r.AuthUsername != nil || r.AuthPassword != nil
}

func (r clusterRequest) applyTo(cluster *model.Cluster) {
	if r.Name != nil {
		cluster.Name = strings.TrimSpace(*r.Name)
	}
	if r.Servers != nil {
		cluster.Servers = strings.TrimSpace(*r.Servers)
	}
	if r.SecurityProtocol != nil {
		cluster.SecurityProtocol = strings.TrimSpace(*r.SecurityProtocol)
	}
	if r.SaslMechanism != nil {
		cluster.SaslMechanism = strings.TrimSpace(*r.SaslMechanism)
	}
	if r.SaslUsername != nil {
		cluster.SaslUsername = strings.TrimSpace(*r.SaslUsername)
	}
	if r.SaslPassword != nil {
		cluster.SaslPassword = *r.SaslPassword
	}
	if r.AuthUsername != nil {
		cluster.SaslUsername = strings.TrimSpace(*r.AuthUsername)
	}
	if r.AuthPassword != nil {
		cluster.SaslPassword = *r.AuthPassword
	}
}

// DeleteCluster deletes a cluster
func (c *ClusterController) DeleteCluster(ctx *gin.Context) {
	ids, err := parseClusterIDs(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, dto.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid cluster ID: " + err.Error(),
		})
		return
	}

	for _, id := range ids {
		if err := c.clusterService.Delete(id); err != nil {
			ctx.JSON(http.StatusInternalServerError, dto.Response{
				Code:    http.StatusInternalServerError,
				Message: "Failed to delete cluster: " + err.Error(),
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Code:    http.StatusOK,
		Message: "Cluster deleted successfully",
	})
}

func parseClusterIDs(value string) ([]uint, error) {
	parts := strings.Split(value, ",")
	ids := make([]uint, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := strconv.ParseUint(part, 10, 32)
		if err != nil {
			return nil, err
		}
		ids = append(ids, uint(id))
	}
	if len(ids) == 0 {
		return nil, strconv.ErrSyntax
	}
	return ids, nil
}

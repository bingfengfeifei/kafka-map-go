package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bingfengfeifei/kafka-map-go/internal/config"
	"github.com/bingfengfeifei/kafka-map-go/internal/controller"
	"github.com/bingfengfeifei/kafka-map-go/internal/middleware"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/service"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
	"github.com/bingfengfeifei/kafka-map-go/pkg/database"
	"github.com/gin-gonic/gin"
)

//go:embed web/index.html web/assets
var webFS embed.FS

func main() {
	// Load configuration
	if err := config.Load("config/config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.InitDB(config.GlobalConfig.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	clusterRepo := repository.NewClusterRepository(db)

	// Initialize utilities
	kafkaManager := util.NewKafkaClientManager()
	tokenCache := util.NewTokenCache(
		time.Duration(config.GlobalConfig.Cache.TokenExpiration)*time.Second,
		config.GlobalConfig.Cache.MaxTokens,
	)

	// Initialize services
	userService := service.NewUserService(userRepo)
	topicStatsRepo := repository.NewTopicStatsRepository(db)
	brokerService := service.NewBrokerService(clusterRepo, kafkaManager)
	topicService := service.NewTopicService(clusterRepo, topicStatsRepo, kafkaManager)
	consumerGroupService := service.NewConsumerGroupService(clusterRepo, kafkaManager)
	clusterService := service.NewClusterService(clusterRepo, kafkaManager, topicService, brokerService, consumerGroupService)

	// Start topic stats background task (refresh every 1 minute)
	topicStatsTask := service.NewTopicStatsTask(clusterRepo, topicStatsRepo, kafkaManager, 1*time.Minute)
	topicStatsTask.Start()

	// Bootstrap clusters provided via configuration/environment variables
	if len(config.GlobalConfig.BootstrapClusters) > 0 {
		if err := clusterService.BootstrapClusters(config.GlobalConfig.BootstrapClusters); err != nil {
			log.Fatalf("Failed to bootstrap clusters: %v", err)
		}
	}

	// Refresh stats immediately after clusters are ready
	log.Println("[Main] Refreshing topic stats for all clusters...")
	topicStatsTask.RefreshAllSync()
	log.Println("[Main] Topic stats refresh complete")

	// Now start the background task for periodic refresh
	// The task is already started, subsequent refreshes will happen every minute

	// Initialize default user
	if err := userService.InitUser(); err != nil {
		log.Fatalf("Failed to initialize default user: %v", err)
	}

	// Initialize controllers
	accountController := controller.NewAccountController(userService, tokenCache)
	clusterController := controller.NewClusterController(clusterService)
	brokerController := controller.NewBrokerController(brokerService)
	topicController := controller.NewTopicController(topicService)
	consumerGroupController := controller.NewConsumerGroupController(consumerGroupService)

	// Setup Gin router
	router := gin.Default()

	// Apply CORS middleware
	router.Use(middleware.CORSMiddleware())

	// API routes with /api prefix
	api := router.Group("/api")
	{
		// Public routes (no authentication required)
		api.POST("/login", accountController.Login)

		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(tokenCache))
		{
			// Account routes
			protected.POST("/logout", accountController.Logout)
			protected.GET("/info", accountController.GetInfo)
			protected.POST("/change-password", accountController.ChangePassword)

			// Cluster routes
			protected.GET("/clusters", clusterController.GetClusters)
			protected.GET("/clusters/paging", clusterController.GetClustersPaged)
			protected.GET("/clusters/:id", clusterController.GetCluster)
			protected.POST("/clusters", clusterController.CreateCluster)
			protected.PUT("/clusters/:id", clusterController.UpdateCluster)
			protected.DELETE("/clusters/:id", clusterController.DeleteCluster)

			// Broker routes
			protected.GET("/brokers", brokerController.GetBrokers)
			protected.GET("/brokers/:id/configs", brokerController.GetBrokerConfigs)
			protected.PUT("/brokers/:id/configs", brokerController.UpdateBrokerConfigs)

			// Topic routes
			protected.GET("/topics", topicController.GetTopics)
			protected.GET("/topics/names", topicController.GetTopicNames)
			protected.GET("/topics/:topic", topicController.GetTopicDetail)
			protected.GET("/topics/:topic/partitions", topicController.GetTopicPartitions)
			protected.GET("/topics/:topic/brokers", topicController.GetTopicBrokers)
			protected.POST("/topics", topicController.CreateTopic)
			protected.POST("/topics/batch-delete", topicController.DeleteTopics)
			protected.POST("/topics/:topic/partitions", topicController.ExpandPartitions)
			protected.GET("/topics/:topic/configs", topicController.GetTopicConfigs)
			protected.PUT("/topics/:topic/configs", topicController.UpdateTopicConfigs)
			protected.GET("/topics/:topic/data", topicController.GetMessages)
			protected.POST("/topics/:topic/data", topicController.SendMessage)

			// Consumer Group routes
			protected.GET("/consumerGroups", consumerGroupController.GetConsumerGroups)
			protected.GET("/consumerGroups/:groupId", consumerGroupController.GetConsumerGroupDetail)
			protected.GET("/consumerGroups/:groupId/describe", consumerGroupController.DescribeConsumerGroup)
			protected.DELETE("/consumerGroups/:groupId", consumerGroupController.DeleteConsumerGroup)
			protected.GET("/topics/:topic/consumerGroups", topicController.GetTopicConsumerGroups)
			protected.GET("/topics/:topic/consumerGroups/:groupId/offset", consumerGroupController.GetConsumerGroupOffset)
			protected.PUT("/topics/:topic/consumerGroups/:groupId/offset", consumerGroupController.ResetConsumerGroupOffset)
		}
	}

	// API routes without /api prefix (for direct API access)
	apiDirect := router.Group("")
	{
		// Public routes (no authentication required)
		apiDirect.POST("/login", accountController.Login)

		// Protected routes (authentication required)
		protected := apiDirect.Group("")
		protected.Use(middleware.AuthMiddleware(tokenCache))
		{
			// Account routes
			protected.POST("/logout", accountController.Logout)
			protected.GET("/info", accountController.GetInfo)
			protected.POST("/change-password", accountController.ChangePassword)

			// Cluster routes
			protected.GET("/clusters", clusterController.GetClusters)
			protected.GET("/clusters/paging", clusterController.GetClustersPaged)
			protected.GET("/clusters/:id", clusterController.GetCluster)
			protected.POST("/clusters", clusterController.CreateCluster)
			protected.PUT("/clusters/:id", clusterController.UpdateCluster)
			protected.DELETE("/clusters/:id", clusterController.DeleteCluster)

			// Broker routes
			protected.GET("/brokers", brokerController.GetBrokers)
			protected.GET("/brokers/:id/configs", brokerController.GetBrokerConfigs)
			protected.PUT("/brokers/:id/configs", brokerController.UpdateBrokerConfigs)

			// Topic routes
			protected.GET("/topics", topicController.GetTopics)
			protected.GET("/topics/names", topicController.GetTopicNames)
			protected.GET("/topics/:topic", topicController.GetTopicDetail)
			protected.GET("/topics/:topic/partitions", topicController.GetTopicPartitions)
			protected.GET("/topics/:topic/brokers", topicController.GetTopicBrokers)
			protected.POST("/topics", topicController.CreateTopic)
			protected.POST("/topics/batch-delete", topicController.DeleteTopics)
			protected.POST("/topics/:topic/partitions", topicController.ExpandPartitions)
			protected.GET("/topics/:topic/configs", topicController.GetTopicConfigs)
			protected.PUT("/topics/:topic/configs", topicController.UpdateTopicConfigs)
			protected.GET("/topics/:topic/data", topicController.GetMessages)
			protected.POST("/topics/:topic/data", topicController.SendMessage)

			// Consumer Group routes
			protected.GET("/consumerGroups", consumerGroupController.GetConsumerGroups)
			protected.GET("/consumerGroups/:groupId", consumerGroupController.GetConsumerGroupDetail)
			protected.GET("/consumerGroups/:groupId/describe", consumerGroupController.DescribeConsumerGroup)
			protected.DELETE("/consumerGroups/:groupId", consumerGroupController.DeleteConsumerGroup)
			protected.GET("/topics/:topic/consumerGroups", topicController.GetTopicConsumerGroups)
			protected.GET("/topics/:topic/consumerGroups/:groupId/offset", consumerGroupController.GetConsumerGroupOffset)
			protected.PUT("/topics/:topic/consumerGroups/:groupId/offset", consumerGroupController.ResetConsumerGroupOffset)
		}
	}

	// Serve embedded static files
	// Serve assets with proper content types
	router.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		fullPath := "web/assets" + filepath

		data, err := webFS.ReadFile(fullPath)
		if err != nil {
			c.String(http.StatusNotFound, "File not found")
			return
		}

		// Set proper content type based on file extension
		contentType := "application/octet-stream"
		if len(filepath) > 3 && filepath[len(filepath)-3:] == ".js" {
			contentType = "application/javascript; charset=utf-8"
		} else if len(filepath) > 4 && filepath[len(filepath)-4:] == ".css" {
			contentType = "text/css; charset=utf-8"
		} else if len(filepath) > 5 && filepath[len(filepath)-5:] == ".json" {
			contentType = "application/json; charset=utf-8"
		} else if len(filepath) > 4 && filepath[len(filepath)-4:] == ".png" {
			contentType = "image/png"
		} else if len(filepath) > 4 && filepath[len(filepath)-4:] == ".jpg" || len(filepath) > 5 && filepath[len(filepath)-5:] == ".jpeg" {
			contentType = "image/jpeg"
		} else if len(filepath) > 4 && filepath[len(filepath)-4:] == ".svg" {
			contentType = "image/svg+xml"
		}

		// Set Content-Type header explicitly before writing data
		c.Header("Content-Type", contentType)
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(data)
	})

	// Serve index.html for root and all non-API routes
	router.GET("/", func(c *gin.Context) {
		data, err := webFS.ReadFile("web/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	router.NoRoute(func(c *gin.Context) {
		// If it's an API route, return 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Not found"})
			return
		}
		// Otherwise serve index.html for SPA routing
		data, err := webFS.ReadFile("web/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load index.html")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	})

	// Start server
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	log.Printf("Starting Kafka-Map server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

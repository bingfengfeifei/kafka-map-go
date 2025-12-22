package repository

import (
	"log"

	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"gorm.io/gorm"
)

type TopicStatsRepository struct {
	db *gorm.DB
}

func NewTopicStatsRepository(db *gorm.DB) *TopicStatsRepository {
	return &TopicStatsRepository{db: db}
}

func (r *TopicStatsRepository) Upsert(stats *model.TopicStats) error {
	return r.db.Model(&model.TopicStats{}).
		Where("cluster_id = ? AND name = ?", stats.ClusterID, stats.Name).
		Assign(map[string]interface{}{
			"total_messages": stats.TotalMessages,
			"last_timestamp": stats.LastTimestamp,
			"updated_at":     stats.UpdatedAt,
		}).FirstOrCreate(stats).Error
}

func (r *TopicStatsRepository) FindByClusterAndName(clusterID uint, name string) (*model.TopicStats, error) {
	var stats model.TopicStats
	err := r.db.Where("cluster_id = ? AND name = ?", clusterID, name).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *TopicStatsRepository) FindByCluster(clusterID uint) ([]model.TopicStats, error) {
	var stats []model.TopicStats
	log.Printf("[TopicStatsRepo] FindByCluster: querying cluster_id=%d", clusterID)
	err := r.db.Where("cluster_id = ?", clusterID).Find(&stats).Error
	log.Printf("[TopicStatsRepo] FindByCluster: found %d records", len(stats))
	return stats, err
}

func (r *TopicStatsRepository) DeleteByClusterAndName(clusterID uint, name string) error {
	return r.db.Where("cluster_id = ? AND name = ?", clusterID, name).Delete(&model.TopicStats{}).Error
}

func (r *TopicStatsRepository) DeleteByCluster(clusterID uint) error {
	return r.db.Where("cluster_id = ?", clusterID).Delete(&model.TopicStats{}).Error
}

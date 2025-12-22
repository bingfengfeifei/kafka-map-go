package model

import (
	"time"

	"gorm.io/gorm"
)

type TopicStats struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	ClusterID     uint           `gorm:"index" json:"clusterId"`
	Name          string         `gorm:"index" json:"name"`
	TotalMessages int64          `json:"totalMessages"`
	LastTimestamp int64          `json:"lastTimestamp"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (TopicStats) TableName() string {
	return "topic_stats"
}

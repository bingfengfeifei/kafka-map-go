package model

import (
	"time"

	"gorm.io/gorm"
)

type Cluster struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Name           string         `gorm:"not null" json:"name"`
	Servers        string         `gorm:"not null" json:"servers"` // Comma-separated broker addresses
	SecurityProtocol string       `json:"securityProtocol"`        // PLAINTEXT, SASL_PLAINTEXT, SASL_SSL, SSL
	SaslMechanism  string         `json:"saslMechanism"`           // PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
	SaslUsername   string         `json:"saslUsername"`
	SaslPassword   string         `json:"saslPassword"`
}

func (Cluster) TableName() string {
	return "clusters"
}

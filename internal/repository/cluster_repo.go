package repository

import (
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"gorm.io/gorm"
)

type ClusterRepository struct {
	db *gorm.DB
}

func NewClusterRepository(db *gorm.DB) *ClusterRepository {
	return &ClusterRepository{db: db}
}

func (r *ClusterRepository) Create(cluster *model.Cluster) error {
	return r.db.Create(cluster).Error
}

func (r *ClusterRepository) FindByID(id uint) (*model.Cluster, error) {
	var cluster model.Cluster
	err := r.db.First(&cluster, id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterRepository) FindAll() ([]model.Cluster, error) {
	var clusters []model.Cluster
	err := r.db.Find(&clusters).Error
	return clusters, err
}

func (r *ClusterRepository) FindAllPaged(pageIndex, pageSize int, name string) ([]model.Cluster, int64, error) {
	var clusters []model.Cluster
	var total int64

	query := r.db.Model(&model.Cluster{})

	// Apply name filter if provided
	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// Count total
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (pageIndex - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Find(&clusters).Error
	return clusters, total, err
}

func (r *ClusterRepository) Update(cluster *model.Cluster) error {
	return r.db.Save(cluster).Error
}

func (r *ClusterRepository) Delete(id uint) error {
	return r.db.Delete(&model.Cluster{}, id).Error
}

func (r *ClusterRepository) ExistsByID(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Cluster{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

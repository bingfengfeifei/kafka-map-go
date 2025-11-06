package service

import (
	"errors"
	"fmt"

	"github.com/bingfengfeifei/kafka-map-go/internal/config"
	"github.com/bingfengfeifei/kafka-map-go/internal/model"
	"github.com/bingfengfeifei/kafka-map-go/internal/repository"
	"github.com/bingfengfeifei/kafka-map-go/internal/util"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// InitUser initializes the default admin user if no users exist
func (s *UserService) InitUser() error {
	count, err := s.userRepo.Count()
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if count == 0 {
		// Create default admin user
		hashedPassword, err := util.HashPassword(config.GlobalConfig.Default.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		user := &model.User{
			Username: config.GlobalConfig.Default.Username,
			Password: hashedPassword,
		}

		if err := s.userRepo.Create(user); err != nil {
			return fmt.Errorf("failed to create default user: %w", err)
		}
	}

	return nil
}

// Authenticate validates username and password
func (s *UserService) Authenticate(username, password string) (*model.User, error) {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	if !util.CheckPassword(user.Password, password) {
		return nil, errors.New("invalid username or password")
	}

	return user, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(username, oldPassword, newPassword string) error {
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return err
	}

	if !util.CheckPassword(user.Password, oldPassword) {
		return errors.New("old password is incorrect")
	}

	hashedPassword, err := util.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.Password = hashedPassword
	return s.userRepo.Update(user)
}

// GetUserByUsername retrieves user by username
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	return s.userRepo.FindByUsername(username)
}

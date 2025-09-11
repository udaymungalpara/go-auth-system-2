package storage

import "go-auth-system/src/models"

type Storage interface {
    GetUserByID(id string) (*models.User, error)
    UpdateUser(user *models.User) error
}
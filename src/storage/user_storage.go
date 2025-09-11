package storage

import "go-auth-system/src/models"

type UserStorage interface {
    CreateUser(user *models.User) error
    GetUserByUsername(username string) (*models.User, error)
    UpdateUser(user *models.User) error
}
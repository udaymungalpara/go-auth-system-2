package handlers

import (
	"go-auth-system/src/models"
	//"go-auth-system/src/storage"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	storage *gorm.DB
}

func NewUserHandler(storage *gorm.DB) *UserHandler {
	return &UserHandler{storage: storage}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if err := h.storage.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	val, _ := strconv.ParseUint(userID, 10, 64)

	// Then cast to uint (platform-dependent size)

	user.ID = uint(val)
	if err := h.storage.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

package models

import (
	"go-auth-system/src/utils"
	"time"
)

type User struct {
	ID               uint   `gorm:"primaryKey"`
	Email            string `gorm:"uniqueIndex;not null"`
	PasswordHash     string `gorm:"not null"`
	FirstName        string
	LastName         string
	IsEmailVerified  bool
	EmailVerifiedAt  *time.Time
	FailedLoginCount int
	LockedUntil      *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	LastLoginAt      *time.Time
}

type RefreshToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
}

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
}

type EmailVerificationToken struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	CreatedAt time.Time `json:"created_at"`
	User      User      `gorm:"foreignKey:UserID" json:"user"`
}

func (u *User) SetPassword(password string) error {
	hash, err := utils.HashPassword(password)
	if err != nil {
		return err
	}
	u.PasswordHash = hash
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return utils.CheckPasswordHash(password, u.PasswordHash)
}

func (u *User) IsAccountLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.LockedUntil)
}

func (u *User) LockAccount(duration time.Duration) {
	lockUntil := time.Now().Add(duration)
	u.LockedUntil = &lockUntil
}

func (u *User) UnlockAccount() {
	u.LockedUntil = nil
	u.FailedLoginCount = 0
}

func (u *User) IncrementFailedLogin() {
	u.FailedLoginCount++
	if u.FailedLoginCount >= 5 {
		// Lock account for 15 minutes after 5 failed attempts
		u.LockAccount(15 * time.Minute)
	}
}

func (u *User) ResetFailedLoginCount() {
	u.FailedLoginCount = 0
	u.UnlockAccount()
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// RegisterResponse represents the response body for user registration
type RegisterResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

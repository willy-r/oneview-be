package model

type User struct {
	ID         uint   `gorm:"primaryKey"`
	Email      string `gorm:"uniqueIndex;not null"`
	Password   string `gorm:"not null"`
	PublicCode string `gorm:"uniqueIndex;not null"`
}

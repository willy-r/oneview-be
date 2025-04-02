package model

import "time"

type Message struct {
	ID         uint   `gorm:"primaryKey"`
	SenderID   uint   `gorm:"not null"`
	ReceiverID uint   `gorm:"not null"`
	Content    string `gorm:"not null"`
	CreatedAt  time.Time
	ReadAt     *time.Time
}

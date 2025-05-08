package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint      `gorm:"primaryKey,autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	PackageID uint    `gorm:"not null,index,index:idx_msg_package,uniqueIndex:message_name_unique"`
	Package   Package `gorm:"constraint:OnDelete:CASCADE,references:PackageID"`
	Name      string  `gorm:"not null,index:idx_msg_package,uniqueIndex:message_name_unique"`

	LatestVersionID *uint           `gorm:"not null,index"`
	LatestVersion   *MessageVersion `gorm:"constraint:OnDelete:SET NULL,references:LatestVersionID"`

	ProtoBody string `gorm:"not null"`
}

// ListMessages lists messages from the database
func ListMessages(db *gorm.DB, limit, offset int, searchTerm string) ([]Message, error) {
	var messages []Message

	query := db.Model(&Message{}).Preload("LatestVersion").Preload("Package")

	// If search term is provided, filter messages by name
	if searchTerm != "" {
		query = query.Where("name LIKE ?", "%"+searchTerm+"%")
	}

	// Order by updated at
	query = query.Order("updated_at DESC")

	// Fetch results
	err := query.Offset(offset).Limit(limit).Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	return messages, nil
}

// CountMessages counts the number of messages in the database
func CountMessages(db *gorm.DB, searchTerm string) (int64, error) {
	var count int64

	query := db.Model(&Message{})

	if searchTerm != "" {
		query = query.Where("name LIKE ?", "%"+searchTerm+"%")
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

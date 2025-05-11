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

func CountMessagesByPackage(db *gorm.DB, packageID uint) (int64, error) {
	var count int64

	query := db.Model(&Message{}).Where("package_id = ?", packageID)

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count messages: %w", err)
	}

	return count, nil
}

// AssignLatestVersion will find all the messages for a given package and assign the latest version to the message
func AssignLatestVersion(db *gorm.DB, packageID uint) error {
	// Fetch all messages for the package
	var messages []Message
	err := db.Model(&Message{}).Where("package_id = ?", packageID).Find(&messages).Error
	if err != nil {
		return fmt.Errorf("failed to assign latest version: %w", err)
	}

	for _, message := range messages {
		// Fetch the latest version for the message
		var msgVersions []MessageVersion
		err = db.Model(&MessageVersion{}).Where("message_id = ?", message.ID).Find(&msgVersions).Error
		if err != nil {
			return fmt.Errorf("failed to assign latest version: %w", err)
		}

		if len(msgVersions) > 0 {
			message.LatestVersionID = &msgVersions[0].ID
		} else {
			message.LatestVersionID = nil
		}

		// Save the message
		err = db.Save(&message).Error
		if err != nil {
			return fmt.Errorf("failed to assign latest version: %w", err)
		}
	}

	return nil
}

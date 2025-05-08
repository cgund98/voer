package db

import (
	"time"

	"gorm.io/gorm"
)

type MessageVersion struct {
	ID        uint      `gorm:"primaryKey,autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	MessageID uint    `gorm:"not null,index,uniqueIndex:message_version_unique"`
	Message   Message `gorm:"constraint:OnDelete:CASCADE,references:MessageID"`

	Version int `gorm:"not null,uniqueIndex:message_version_unique"`

	ProtoBody        string `gorm:"not null"`
	SerializedSchema string `gorm:"not null"`
}

// GetNextMessageVersion returns the next message version for a given message ID
func GetNextMessageVersion(db *gorm.DB, messageID uint) (int, error) {
	var messageVersions []MessageVersion
	result := db.Where("message_id = ?", messageID).Order("version DESC").Limit(1).Find(&messageVersions)
	if result.Error != nil {
		return 0, result.Error
	}

	if len(messageVersions) == 0 {
		return 1, nil
	}

	latestVersion := messageVersions[0]

	return latestVersion.Version + 1, nil
}

package db

import "time"

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

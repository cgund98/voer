package db

import "time"

type Message struct {
	ID        uint      `gorm:"primaryKey,autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	PackageID uint `gorm:"not null,index"`
	// Package   Package `gorm:"constraint:OnDelete:CASCADE"`

	ProtoBody string `gorm:"not null"`
}

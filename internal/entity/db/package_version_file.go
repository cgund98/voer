package db

import "time"

// PackageVersion is the database model for a package version entity
type PackageVersionFile struct {
	ID        uint      `gorm:"primaryKey,autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	PackageVersionID uint           `gorm:"not null,index"`
	PackageVersion   PackageVersion `gorm:"constraint:OnDelete:CASCADE,foreignKey:PackageVersionID,references:ID"`

	FileName     string `gorm:"not null"`
	FileContents string `gorm:"not null"`
}

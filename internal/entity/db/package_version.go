package db

import "time"

// PackageVersion is the database model for a package version entity
type PackageVersion struct {
	ID        uint      `gorm:"primaryKey,autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	PackageID uint `gorm:"not null,index,uniqueIndex:package_version_number_unique"`
	Version   int  `gorm:"not null,uniqueIndex:package_version_number_unique"`

	Package Package              `gorm:"constraint:OnDelete:CASCADE,foreignKey:PackageID,references:ID"`
	Files   []PackageVersionFile `gorm:"constraint:OnDelete:CASCADE,foreignKey:PackageVersionID,references:ID"`
}

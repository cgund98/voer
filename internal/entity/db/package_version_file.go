package db

import (
	"time"

	"gorm.io/gorm"
)

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

func ListPackageVersionFiles(db *gorm.DB, packageVersionID uint) ([]PackageVersionFile, error) {
	var files []PackageVersionFile
	if err := db.Where("package_version_id = ?", packageVersionID).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

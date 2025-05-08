package db

import (
	"time"

	"gorm.io/gorm"
)

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

// GetNextPackageVersion returns the next package version for a given package ID
func GetNextPackageVersion(db *gorm.DB, packageID uint) (int, error) {
	var pkgVersions []PackageVersion
	result := db.Where("package_id = ?", packageID).Order("version DESC").Limit(1).Find(&pkgVersions)
	if result.Error != nil {
		return 0, result.Error
	}

	if len(pkgVersions) == 0 {
		return 1, nil
	}

	latestVersion := pkgVersions[0]

	return latestVersion.Version + 1, nil
}

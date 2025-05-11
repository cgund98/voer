package db

import (
	"fmt"
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

	MessageVersions []MessageVersion `gorm:"constraint:OnDelete:CASCADE,foreignKey:PackageVersionID,references:ID"`
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

func ListPackageVersions(db *gorm.DB, packageID uint) ([]PackageVersion, error) {
	var pkgVersions []PackageVersion
	result := db.Where("package_id = ?", packageID).Order("version DESC").Find(&pkgVersions)
	if result.Error != nil {
		return nil, result.Error
	}

	return pkgVersions, nil
}

func DeletePackageVersion(db *gorm.DB, packageVersionID uint) error {
	// Get the package version
	var packageVersion PackageVersion
	result := db.First(&packageVersion, packageVersionID)
	if result.Error != nil {
		return fmt.Errorf("failed to get package version: %w", result.Error)
	}

	// Get related package
	var pkg Package
	result = db.First(&pkg, packageVersion.PackageID)
	if result.Error != nil {
		return fmt.Errorf("failed to get package: %w", result.Error)
	}

	result = db.Delete(&packageVersion, packageVersionID)
	if result.Error != nil {
		return fmt.Errorf("failed to delete package version: %w", result.Error)
	}

	// Update the package's latest version
	var pkgVersions []PackageVersion
	result = db.Where("package_id = ?", packageVersion.PackageID).Order("version DESC").Limit(1).Find(&pkgVersions)
	if result.Error != nil {
		return fmt.Errorf("failed to get latest package version: %w", result.Error)
	}

	var latestVersion *PackageVersion
	if len(pkgVersions) > 0 {
		latestVersion = &pkgVersions[0]
	}

	if latestVersion != nil {
		pkg.LatestVersionID = &latestVersion.ID
		result = db.Save(&pkg)
		if result.Error != nil {
			return fmt.Errorf("failed to update package latest version: %w", result.Error)
		}
	}

	return nil
}

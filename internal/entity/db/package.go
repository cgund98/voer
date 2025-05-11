package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Package is the database model for a package entity
type Package struct {
	ID        uint      `gorm:"primaryKey,autoIncrement"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`

	PackageName     string `gorm:"unique,index"`
	LatestVersionID *uint  `gorm:"index"`

	LatestVersion *PackageVersion  `gorm:"foreignKey:ID;references:LatestVersionID"`
	Versions      []PackageVersion `gorm:"constraint:OnDelete:CASCADE,foreignKey:PackageID,references:ID"`
}

func ListPackages(db *gorm.DB, limit, offset int, searchTerm string) ([]Package, error) {
	var packages []Package

	query := db.Model(&Package{}).Preload("LatestVersion")

	if searchTerm != "" {
		query = query.Where("package_name LIKE ?", "%"+searchTerm+"%")
	}

	err := query.Offset(offset).Limit(limit).Find(&packages).Error
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	return packages, nil
}

func CountPackages(db *gorm.DB, searchTerm string) (int64, error) {
	var count int64

	query := db.Model(&Package{})

	if searchTerm != "" {
		query = query.Where("package_name LIKE ?", "%"+searchTerm+"%")
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count packages: %w", err)
	}

	return count, nil
}

func GetPackage(db *gorm.DB, id uint64) (*Package, error) {
	var pkg Package
	err := db.Model(&Package{}).Preload("LatestVersion").Where("id = ?", id).First(&pkg).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get package: %w", err)
	}

	return &pkg, nil
}

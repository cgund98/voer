package db

import "time"

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

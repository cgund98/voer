package grpc

import (
	"context"

	v1 "github.com/cgund98/voer/api/v1"
	"github.com/cgund98/voer/internal/entity/ctrl"
	"gorm.io/gorm"
)

type PackageSvc struct {
	v1.UnimplementedPackageSvcServer

	DB *gorm.DB
}

func NewPackageSvc(db *gorm.DB) *PackageSvc {
	return &PackageSvc{DB: db}
}

func (s *PackageSvc) UploadPackageVersion(ctx context.Context, req *v1.UploadPackageVersionRequest) (*v1.UploadPackageVersionResponse, error) {
	return ctrl.CreatePackageVersion(ctx, s.DB, req)
}

func (s *PackageSvc) ValidatePackageVersion(ctx context.Context, req *v1.ValidatePackageVersionRequest) (*v1.ValidatePackageVersionResponse, error) {
	return ctrl.ValidatePackageVersion(ctx, s.DB, req)
}

func (s *PackageSvc) GetPackageVersion(ctx context.Context, req *v1.GetPackageVersionRequest) (*v1.GetPackageVersionResponse, error) {
	return ctrl.GetPackageVersion(ctx, s.DB, req)
}

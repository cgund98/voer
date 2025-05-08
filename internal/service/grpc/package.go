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

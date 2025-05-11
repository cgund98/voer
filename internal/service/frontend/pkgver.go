package frontend

import (
	"net/http"

	"github.com/cgund98/voer/internal/entity/db"
	"github.com/cgund98/voer/internal/infra/logging"
	"github.com/cgund98/voer/internal/ui/components/pkgver"

	"github.com/ggicci/httpin"
)

type ListPackageVersionsInput struct {
	PackageID uint `in:"query=package_id"`
}

func (s *Service) HandleListPackageVersions(w http.ResponseWriter, r *http.Request) {
	// Parse inputs
	input := r.Context().Value(httpin.Input).(*ListPackageVersionsInput)

	// List package versions
	pkgVers, err := db.ListPackageVersions(s.db, input.PackageID)
	if err != nil {
		logging.Logger.Error("Failed to list Package Versions", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format input
	inputs := []pkgver.PackageVersionTableInput{}
	for _, pkgVer := range pkgVers {
		inputs = append(inputs, pkgver.PackageVersionTableInput{
			PackageVersionID: pkgVer.ID,
			Version:          pkgVer.Version,
			UpdatedAt:        pkgVer.UpdatedAt,
		})
	}

	// Render component
	component := pkgver.PackageVersionTable(inputs)
	err = component.Render(r.Context(), w)
	if err != nil {
		logging.Logger.Error("Failed to render Package Version Table", "error", err)
	}
}

type DeletePackageVersionInput struct {
	PackageVersionID uint `in:"path=package_version_id"`
}

func (s *Service) HandleDeletePackageVersion(w http.ResponseWriter, r *http.Request) {
	// Parse inputs
	input := r.Context().Value(httpin.Input).(*DeletePackageVersionInput)

	// Attempt to fetch the package version
	var pkgVer db.PackageVersion
	err := s.db.Model(&db.PackageVersion{}).Where("id = ?", input.PackageVersionID).First(&pkgVer).Error
	if err != nil {
		logging.Logger.Warn("Failed to get Package Version", "error", err)
		http.Error(w, "Package version not found", http.StatusNotFound)
		return
	}
	// Create transaction
	tx := s.db.Begin()
	if err != nil {
		logging.Logger.Error("Failed to create transaction", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete package version
	err = db.DeletePackageVersion(tx, input.PackageVersionID)
	if err != nil {
		logging.Logger.Error("Failed to delete Package Version", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update messages
	err = db.AssignLatestVersion(tx, pkgVer.PackageID)
	if err != nil {
		logging.Logger.Error("Failed to update messages", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		logging.Logger.Error("Failed to commit transaction", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set HX-Trigger header
	w.Header().Set("HX-Trigger", "package-version-deleted")

	// Respond with text
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Package version deleted successfully"))
	if err != nil {
		logging.Logger.Error("Failed to write response", "error", err)
	}
}

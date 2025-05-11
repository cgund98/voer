package frontend

import (
	"net/http"

	db "github.com/cgund98/voer/internal/entity/db"
	"github.com/cgund98/voer/internal/infra/logging"
	pkgverfile "github.com/cgund98/voer/internal/ui/components/pkgverfile"
	"github.com/ggicci/httpin"
)

type ListPackageVersionFilesInput struct {
	PackageVersionID uint `in:"query=package_version_id"`
}

func (s *Service) HandleListPackageVersionFiles(w http.ResponseWriter, r *http.Request) {
	// Parse inputs
	input := r.Context().Value(httpin.Input).(*ListPackageVersionFilesInput)

	// Fetch package version files
	packageVersionFiles, err := db.ListPackageVersionFiles(s.db, input.PackageVersionID)
	if err != nil {
		logging.Logger.Error("Failed to list package version files", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format response
	cardInputs := []pkgverfile.PackageVersionFileListCardInput{}
	for _, file := range packageVersionFiles {
		cardInputs = append(cardInputs, pkgverfile.PackageVersionFileListCardInput{
			FileName:     file.FileName,
			FileContents: file.FileContents,
		})
	}

	// Render package version files
	component := pkgverfile.PackageVersionFilesList(cardInputs)
	err = component.Render(r.Context(), w)
	if err != nil {
		logging.Logger.Error("Failed to render package version files", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

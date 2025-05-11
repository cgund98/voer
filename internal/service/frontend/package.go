package frontend

import (
	"fmt"
	"net/http"

	"github.com/ggicci/httpin"
	"google.golang.org/protobuf/proto"

	"github.com/cgund98/voer/internal/entity/db"
	"github.com/cgund98/voer/internal/infra/logging"

	msgComponents "github.com/cgund98/voer/internal/ui/components/package"
	page "github.com/cgund98/voer/internal/ui/page"
)

type ListPackagesInput struct {
	Page   int    `in:"query=page"`
	Search string `in:"query=search"`
}

// HandleListPackages handles the list Packages request
func (s *Service) HandleListPackages(w http.ResponseWriter, r *http.Request) {
	// Parse inputs
	input := r.Context().Value(httpin.Input).(*ListPackagesInput)

	// Fetch Packages
	limit := pageSize
	offset := (input.Page - 1) * limit

	// List Packages
	packages, err := db.ListPackages(s.db, limit, offset, input.Search)
	if err != nil {
		logging.Logger.Error("Failed to list Packages", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Count Packages
	count, err := db.CountPackages(s.db, input.Search)
	if err != nil {
		logging.Logger.Error("Failed to count Packages", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Count messages for each package
	msgCounts := make(map[uint]int)
	for _, Package := range packages {
		messageCount, err := db.CountMessagesByPackage(s.db, Package.ID)
		if err != nil {
			logging.Logger.Error("Failed to count Messages", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		msgCounts[Package.ID] = int(messageCount)
	}

	// Set hx-trigger header
	w.Header().Set("HX-Trigger", fmt.Sprintf("{\"package-count\": %d}", count))

	// Calculate next page
	nextPage := proto.Int32(int32(input.Page + 1))
	if len(packages) < int(pageSize) {
		nextPage = nil
	}

	// Format Packages
	cardInputs := make([]msgComponents.PackageCardInput, len(packages))
	for i, Package := range packages {
		msgInput := msgComponents.PackageCardInput{
			PackageName: Package.PackageName,
			PackageID:   Package.ID,
		}
		if Package.LatestVersion != nil {
			msgInput.Version = Package.LatestVersion.Version
			msgInput.UpdatedAt = Package.LatestVersion.UpdatedAt
			msgInput.MessageCount = msgCounts[Package.ID]
		}
		cardInputs[i] = msgInput
	}

	// Render component
	component := msgComponents.CardsList(nextPage, cardInputs)
	err = component.Render(r.Context(), w)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error rendering cards list: %v", err))
	}
}

type PackagePageInput struct {
	PackageID uint64 `in:"path=package_id"`
}

func (s *Service) HandlePackagePage(w http.ResponseWriter, r *http.Request) {
	// Parse inputs
	input := r.Context().Value(httpin.Input).(*PackagePageInput)

	// Fetch Package
	pkg, err := db.GetPackage(s.db, input.PackageID)
	if err != nil {
		logging.Logger.Error("Failed to get Package", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format input
	pageInput := page.PackagePageInput{
		PackageID:       pkg.ID,
		PackageName:     pkg.PackageName,
		LatestVersionID: pkg.LatestVersionID,
	}

	if pkg.LatestVersion != nil {
		pageInput.PackageVersion = &pkg.LatestVersion.Version
		pageInput.PackageUpdatedAt = &pkg.LatestVersion.UpdatedAt
	}

	// Count messages
	messageCount, err := db.CountMessagesByPackage(s.db, pkg.ID)
	if err != nil {
		logging.Logger.Error("Failed to count Messages", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pageInput.PackageMessageCount = int(messageCount)

	// Render component
	component := page.PackagePage(pageInput)
	err = component.Render(r.Context(), w)
	if err != nil {
		logging.Logger.Error(fmt.Sprintf("Error rendering package page: %v", err))
	}
}

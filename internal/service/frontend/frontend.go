package frontend

import (
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/ggicci/httpin"
	httpin_integration "github.com/ggicci/httpin/integration"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator"
	slogchi "github.com/samber/slog-chi"
	"gorm.io/gorm"

	"github.com/cgund98/voer/internal/infra/config"
	"github.com/cgund98/voer/internal/infra/logging"
	"github.com/cgund98/voer/internal/ui/page"
)

type Service struct {
	config    *config.Config
	router    chi.Router
	validator *validator.Validate

	db *gorm.DB
}

func NewService(config *config.Config, db *gorm.DB) *Service {
	return &Service{
		config:    config,
		router:    chi.NewRouter(),
		validator: validator.New(),
		db:        db,
	}
}

// Init will initiate all routes
func (fe *Service) Init() {

	// Register a directive named "path" to retrieve values from `chi.URLParam`,
	// i.e. decode path variables.
	httpin_integration.UseGochiURLParam("path", chi.URLParam)

	// Middleware
	fe.router.Use(slogchi.New(logging.Logger))
	fe.router.Use(middleware.Recoverer)

	fe.router.Handle("/*", templ.Handler(page.NotFoundPage()))

	// Routes
	fe.router.Handle("/", templ.Handler(page.Messages()))
	fe.router.Handle("/view/packages", templ.Handler(page.Packages()))
	fe.router.With(httpin.NewInput(PackagePageInput{})).Get("/view/packages/{package_id}", http.HandlerFunc(fe.HandlePackagePage))

	fe.router.With(httpin.NewInput(ListMessagesInput{})).Get("/messages", http.HandlerFunc(fe.HandleListMessages))
	fe.router.With(httpin.NewInput(ListPackagesInput{})).Get("/packages", http.HandlerFunc(fe.HandleListPackages))
	fe.router.With(httpin.NewInput(ListPackageVersionFilesInput{})).Get("/packages-version-files", http.HandlerFunc(fe.HandleListPackageVersionFiles))

	fe.router.With(httpin.NewInput(ListPackageVersionsInput{})).Get("/packages-versions", http.HandlerFunc(fe.HandleListPackageVersions))
	fe.router.With(httpin.NewInput(DeletePackageVersionInput{})).Delete("/packages-versions/{package_version_id}", http.HandlerFunc(fe.HandleDeletePackageVersion))

	// static files
	fe.router.Handle("/static/app.css", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		if _, err := w.Write([]byte(css)); err != nil {
			logging.Logger.Error("Failed to write CSS", "error", err)
		}
	}))

	// Health Check
	fe.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Service is healthy.")); err != nil {
			logging.Logger.Error("Failed to write health check", "error", err)
		}
	})
}

// Start will listen and serve on a given port
func (o *Service) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	logging.Logger.Info("Starting frontend service...", "address", addr)
	return http.ListenAndServe(addr, o.router)
}

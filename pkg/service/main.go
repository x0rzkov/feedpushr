package service

import (
	"context"
	"net/http"

	"github.com/goadesign/goa"
	"github.com/goadesign/goa/middleware"
	"github.com/ncarlier/feedpushr/autogen/app"
	"github.com/ncarlier/feedpushr/pkg/aggregator"
	"github.com/ncarlier/feedpushr/pkg/assets"
	"github.com/ncarlier/feedpushr/pkg/config"
	"github.com/ncarlier/feedpushr/pkg/controller"
	"github.com/ncarlier/feedpushr/pkg/filter"
	"github.com/ncarlier/feedpushr/pkg/logging"
	"github.com/ncarlier/feedpushr/pkg/opml"
	"github.com/ncarlier/feedpushr/pkg/output"
	"github.com/ncarlier/feedpushr/pkg/plugin"
	"github.com/ncarlier/feedpushr/pkg/store"
	"github.com/rs/zerolog/log"
)

// Service is the global service
type Service struct {
	db         store.DB
	srv        *goa.Service
	aggregator *aggregator.Manager
}

// ClearCache clear DB cache
func (s *Service) ClearCache() error {
	return s.db.ClearCache()
}

// ImportOPMLFile imports OPML file
func (s *Service) ImportOPMLFile(filename string) error {
	o, err := opml.NewOPMLFromFile(filename)
	if err != nil {
		return err
	}
	err = opml.ImportOPMLToDB(o, s.db)
	if err != nil {
		log.Error().Err(err)
	}
	return nil
}

// ListenAndServe starts server
func (s *Service) ListenAndServe(ListenAddr string) error {
	if err := s.aggregator.Start(); err != nil {
		return err
	}
	if err := s.srv.ListenAndServe(ListenAddr); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown service
func (s *Service) Shutdown(ctx context.Context) error {
	s.aggregator.Shutdown()
	s.srv.CancelAll()
	s.srv.Server.SetKeepAlivesEnabled(false)
	return s.srv.Server.Shutdown(ctx)
}

// Configure the global service
func Configure(db store.DB, conf config.Config) (*Service, error) {
	// Load plugins
	pr, err := plugin.NewPluginRegistry(conf.Plugins.Values())
	if err != nil {
		log.Error().Err(err).Msg("unable to init plugins")
		return nil, err
	}
	// Init chain filter
	cf, err := filter.NewChainFilter(conf.Filters.Values(), pr)
	if err != nil {
		log.Error().Err(err).Msg("unable to init filter chain")
		return nil, err
	}

	// Init output manager
	om, err := output.NewManager(db, conf.Outputs.Values(), conf.CacheRetention, pr, cf)
	if err != nil {
		log.Error().Err(err).Msg("unable to init output manager")
		return nil, err
	}

	// Init aggregator daemon
	var callbackURL string
	if conf.PublicURL != "" {
		callbackURL = conf.PublicURL + "/v1/pshb"
	}
	am := aggregator.NewManager(db, om, conf.Delay, conf.Timeout, callbackURL)

	// Create service
	srv := goa.New("feedpushr")

	// Set custom logger
	logger := log.With().Str("component", "server").Logger()
	srv.WithLogger(logging.NewLogAdapter(logger))

	// Mount middleware
	srv.Use(middleware.RequestID())
	srv.Use(middleware.LogRequest(false))
	srv.Use(middleware.ErrorHandler(srv, true))
	srv.Use(middleware.Recover())

	// Mount "feed" controller
	app.MountFeedController(srv, controller.NewFeedController(srv, db, am))
	// Mount "filter" controller
	app.MountFilterController(srv, controller.NewFilterController(srv, cf))
	// Mount "output" controller
	app.MountOutputController(srv, controller.NewOutputController(srv, om))
	// Mount "health" controller
	app.MountHealthController(srv, controller.NewHealthController(srv))
	// Mount "swagger" controller
	app.MountSwaggerController(srv, controller.NewSwaggerController(srv))
	// Mount "opml" controller
	app.MountOpmlController(srv, controller.NewOpmlController(srv, db))
	// Mount "vars" controller
	app.MountVarsController(srv, controller.NewVarsController(srv))
	// Mount "pshb" controller (only if public URL is configured)
	if conf.PublicURL != "" {
		app.MountPshbController(srv, controller.NewPshbController(srv, db, am, om))
	}
	// Mount custom handlers (aka: not generated)...
	srv.Mux.Handle("GET", "/ui/*asset", assets.Handler())
	srv.Mux.Handle("GET", "/ui/", assets.Handler())

	return &Service{
		db:         db,
		srv:        srv,
		aggregator: am,
	}, nil
}

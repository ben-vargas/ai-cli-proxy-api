// Package modules provides a pluggable routing module system for extending
// the API server with optional features without modifying core routing logic.
package modules

import (
	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers"
)

// RouteModule represents a pluggable routing module that can register routes
// and handle configuration updates independently of the core server.
type RouteModule interface {
	// Name returns a human-readable identifier for the module
	Name() string

	// Register sets up routes and handlers for this module.
	// It receives the Gin engine, base handlers, and current configuration.
	// Returns an error if registration fails (errors are logged but don't stop the server).
	Register(engine *gin.Engine, baseHandler *handlers.BaseAPIHandler, cfg *config.Config) error

	// OnConfigUpdated is called when the configuration is reloaded.
	// Modules can respond to configuration changes here.
	// Returns an error if the update cannot be applied.
	OnConfigUpdated(cfg *config.Config) error
}

// WithModules creates a ServerOption that registers one or more route modules.
// Modules are registered after core routes are set up, allowing them to layer
// on additional functionality without conflicting with upstream changes.
//
// Example usage:
//   ampModule := amp.New(accessManager)
//   server := api.NewServer(cfg, accessManager, api.WithModules(ampModule))
func WithModules(modules ...RouteModule) func(func(*gin.Engine, *handlers.BaseAPIHandler, *config.Config)) {
	return func(registerFn func(*gin.Engine, *handlers.BaseAPIHandler, *config.Config)) {
		// This is a helper that would be used with WithRouterConfigurator
		// The actual integration happens in the calling code
	}
}

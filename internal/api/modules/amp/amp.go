// Package amp implements the Amp CLI routing module, providing OAuth-based
// integration with Amp CLI for ChatGPT and Anthropic subscriptions.
package amp

import (
	"fmt"
	"net/http/httputil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/internal/config"
	sdkaccess "github.com/router-for-me/CLIProxyAPI/v6/sdk/access"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers"
	log "github.com/sirupsen/logrus"
)

// AmpModule implements the RouteModule interface for Amp CLI integration.
// It provides:
//   - Reverse proxy to Amp control plane for OAuth/management
//   - Provider-specific route aliases (/api/provider/{provider}/...)
//   - Automatic gzip decompression for misconfigured upstreams
type AmpModule struct {
	secretSource    SecretSource
	proxy           *httputil.ReverseProxy
	accessManager   *sdkaccess.Manager
	authMiddleware_ gin.HandlerFunc
	enabled         bool
}

// New creates a new Amp routing module with the given access manager.
// The authMiddleware function should be the AuthMiddleware from the api package.
func New(accessManager *sdkaccess.Manager, authMiddleware gin.HandlerFunc) *AmpModule {
	return &AmpModule{
		accessManager:   accessManager,
		authMiddleware_: authMiddleware,
	}
}

// Name returns the module identifier
func (m *AmpModule) Name() string {
	return "amp-routing"
}

// Register sets up Amp routes if configured.
// This is called after core routes are established, allowing us to layer
// on Amp functionality without touching setupRoutes().
func (m *AmpModule) Register(engine *gin.Engine, baseHandler *handlers.BaseAPIHandler, cfg *config.Config) error {
	upstreamURL := strings.TrimSpace(cfg.AmpUpstreamURL)
	if upstreamURL == "" {
		log.Debug("Amp routing disabled (no upstream URL configured)")
		m.enabled = false
		return nil
	}

	// Create secret source with precedence: config > env > file
	// Cache secrets for 5 minutes to reduce file I/O
	secretSource := NewMultiSourceSecret(cfg.AmpUpstreamAPIKey, 0 /* default 5min */)
	m.secretSource = secretSource

	// Create reverse proxy with gzip handling via ModifyResponse
	proxy, err := createReverseProxy(upstreamURL, secretSource)
	if err != nil {
		return fmt.Errorf("failed to create amp proxy: %w", err)
	}

	m.proxy = proxy
	m.enabled = true

	// Register routes
	handler := proxyHandler(proxy)
	m.registerManagementRoutes(engine, handler)
	m.registerProviderAliases(engine, baseHandler)

	log.Infof("Amp routing enabled for upstream: %s", upstreamURL)
	return nil
}

// OnConfigUpdated handles configuration updates.
// Currently requires restart for URL changes (could be enhanced for dynamic updates).
func (m *AmpModule) OnConfigUpdated(cfg *config.Config) error {
	if !m.enabled {
		log.Debug("Amp routing not enabled, skipping config update")
		return nil
	}

	upstreamURL := strings.TrimSpace(cfg.AmpUpstreamURL)
	if upstreamURL == "" {
		log.Warn("Amp upstream URL removed from config, restart required to disable")
		return nil
	}

	// If API key changed, invalidate the cache
	if m.secretSource != nil {
		if ms, ok := m.secretSource.(*MultiSourceSecret); ok {
			ms.InvalidateCache()
			log.Debug("Amp secret cache invalidated due to config update")
		}
	}

	log.Debug("Amp config updated (restart required for URL changes)")
	return nil
}

// authMiddleware returns the authentication middleware for provider routes
func (m *AmpModule) authMiddleware() gin.HandlerFunc {
	if m.authMiddleware_ != nil {
		return m.authMiddleware_
	}

	// Fallback: no authentication (should not happen in production)
	log.Warn("Amp module: no auth middleware provided, allowing all requests")
	return func(c *gin.Context) {
		c.Next()
	}
}

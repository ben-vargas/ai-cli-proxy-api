package amp

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers/claude"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers/gemini"
	"github.com/router-for-me/CLIProxyAPI/v6/sdk/api/handlers/openai"
)

// registerManagementRoutes registers Amp management proxy routes
// These routes proxy through to the Amp control plane for OAuth, user management, etc.
func (m *AmpModule) registerManagementRoutes(engine *gin.Engine, proxyHandler gin.HandlerFunc) {
	ampAPI := engine.Group("/api")

	// Management routes - these are proxied directly to Amp upstream
	ampAPI.Any("/internal", proxyHandler)
	ampAPI.Any("/internal/*path", proxyHandler)
	ampAPI.Any("/user", proxyHandler)
	ampAPI.Any("/user/*path", proxyHandler)
	ampAPI.Any("/auth", proxyHandler)
	ampAPI.Any("/auth/*path", proxyHandler)
	ampAPI.Any("/meta", proxyHandler)
	ampAPI.Any("/meta/*path", proxyHandler)
	ampAPI.Any("/ads", proxyHandler)
	ampAPI.Any("/telemetry", proxyHandler)
	ampAPI.Any("/telemetry/*path", proxyHandler)
	ampAPI.Any("/threads", proxyHandler)
	ampAPI.Any("/threads/*path", proxyHandler)
	ampAPI.Any("/otel", proxyHandler)
	ampAPI.Any("/otel/*path", proxyHandler)

	// Google v1beta1 passthrough (Gemini native API)
	ampAPI.Any("/provider/google/v1beta1/*path", proxyHandler)
}

// registerProviderAliases registers /api/provider/{provider}/... routes
// These allow Amp CLI to route requests like:
//   /api/provider/openai/v1/chat/completions
//   /api/provider/anthropic/v1/messages
//   /api/provider/google/v1beta/models
func (m *AmpModule) registerProviderAliases(engine *gin.Engine, baseHandler *handlers.BaseAPIHandler) {
	// Create handler instances for different providers
	openaiHandlers := openai.NewOpenAIAPIHandler(baseHandler)
	geminiHandlers := gemini.NewGeminiAPIHandler(baseHandler)
	claudeCodeHandlers := claude.NewClaudeCodeAPIHandler(baseHandler)
	openaiResponsesHandlers := openai.NewOpenAIResponsesAPIHandler(baseHandler)

	// Provider-specific routes under /api/provider/:provider
	ampProviders := engine.Group("/api/provider")
	ampProviders.Use(m.authMiddleware())

	provider := ampProviders.Group("/:provider")

	// Dynamic models handler - routes to appropriate provider based on path parameter
	ampModelsHandler := func(c *gin.Context) {
		providerName := strings.ToLower(c.Param("provider"))

		switch providerName {
		case "anthropic":
			claudeCodeHandlers.ClaudeModels(c)
		case "google":
			geminiHandlers.GeminiModels(c)
		default:
			// Default to OpenAI-compatible (works for openai, groq, cerebras, etc.)
			openaiHandlers.OpenAIModels(c)
		}
	}

	// Root-level routes (for providers that omit /v1, like groq/cerebras)
	provider.GET("/models", ampModelsHandler)
	provider.POST("/chat/completions", openaiHandlers.ChatCompletions)
	provider.POST("/completions", openaiHandlers.Completions)
	provider.POST("/responses", openaiResponsesHandlers.Responses)

	// /v1 routes (OpenAI/Claude-compatible endpoints)
	v1Amp := provider.Group("/v1")
	{
		v1Amp.GET("/models", ampModelsHandler)

		// OpenAI-compatible endpoints
		v1Amp.POST("/chat/completions", openaiHandlers.ChatCompletions)
		v1Amp.POST("/completions", openaiHandlers.Completions)
		v1Amp.POST("/responses", openaiResponsesHandlers.Responses)

		// Claude/Anthropic-compatible endpoints
		v1Amp.POST("/messages", claudeCodeHandlers.ClaudeMessages)
		v1Amp.POST("/messages/count_tokens", claudeCodeHandlers.ClaudeCountTokens)
	}

	// /v1beta routes (Gemini native API)
	v1betaAmp := provider.Group("/v1beta")
	{
		v1betaAmp.GET("/models", geminiHandlers.GeminiModels)
		v1betaAmp.POST("/models/:action", geminiHandlers.GeminiHandler)
		v1betaAmp.GET("/models/:action", geminiHandlers.GeminiGetHandler)
	}
}



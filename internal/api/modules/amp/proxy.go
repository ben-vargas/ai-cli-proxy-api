package amp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// createReverseProxy creates a reverse proxy handler for Amp upstream
// with automatic gzip decompression via ModifyResponse
func createReverseProxy(upstreamURL string, secretSource SecretSource) (*httputil.ReverseProxy, error) {
	parsed, err := url.Parse(upstreamURL)
	if err != nil {
		return nil, fmt.Errorf("invalid amp upstream url: %w", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(parsed)
	originalDirector := proxy.Director

	// Modify outgoing requests to inject API key and fix routing
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = parsed.Host

		// Preserve correlation headers for debugging
		if req.Header.Get("X-Request-ID") == "" {
			// Could generate one here if needed
		}

		// Inject API key from secret source (precedence: config > env > file)
		if key, err := secretSource.Get(req.Context()); err == nil && key != "" {
			req.Header.Set("X-Api-Key", key)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
		} else if err != nil {
			log.Warnf("amp secret source error (continuing without auth): %v", err)
		}
	}

	// Modify incoming responses to handle gzip without Content-Encoding
	// This addresses the same issue as inline handler gzip handling, but at the proxy level
	proxy.ModifyResponse = func(resp *http.Response) error {
		// Only process successful responses
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil
		}

		// Skip if already marked as gzip (Content-Encoding set)
		if resp.Header.Get("Content-Encoding") != "" {
			return nil
		}

		// Skip streaming responses (SSE, chunked)
		if isStreamingResponse(resp) {
			return nil
		}

		// Peek at first 2 bytes to detect gzip magic bytes
		header := make([]byte, 2)
		n, _ := io.ReadFull(resp.Body, header)
		
		// Check for gzip magic bytes (0x1f 0x8b)
		// If n < 2, we didn't get enough bytes, so it's not gzip
		if n >= 2 && header[0] == 0x1f && header[1] == 0x8b {
			// It's gzip - read the rest of the body
			rest, err := io.ReadAll(resp.Body)
			if err != nil {
				// Restore what we read and return original body
				resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(header[:n]), resp.Body))
				return nil
			}
			
			// Reconstruct complete gzipped data
			gzippedData := append(header[:n], rest...)

			// Decompress
			gzipReader, err := gzip.NewReader(bytes.NewReader(gzippedData))
			if err != nil {
				log.Warnf("amp proxy: gzip header detected but decompress failed: %v", err)
				// Return original gzipped body
				resp.Body = io.NopCloser(bytes.NewReader(gzippedData))
				return nil
			}

			decompressed, err := io.ReadAll(gzipReader)
			_ = gzipReader.Close()
			if err != nil {
				log.Warnf("amp proxy: gzip decompress error: %v", err)
				// Return original gzipped body
				resp.Body = io.NopCloser(bytes.NewReader(gzippedData))
				return nil
			}

			// Replace body with decompressed content
			resp.Body = io.NopCloser(bytes.NewReader(decompressed))
			resp.ContentLength = int64(len(decompressed))

			// Update headers to reflect decompressed state
			resp.Header.Del("Content-Encoding")                                      // No longer compressed
			resp.Header.Del("Content-Length")                                        // Remove stale compressed length
			resp.Header.Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10)) // Set decompressed length

			log.Debugf("amp proxy: decompressed gzip response (%d -> %d bytes)", len(gzippedData), len(decompressed))
		} else {
			// Not gzip - restore original body with peeked bytes
			// Handle edge cases: n might be 0, 1, or 2 depending on EOF
			resp.Body = io.NopCloser(io.MultiReader(bytes.NewReader(header[:n]), resp.Body))
		}

		return nil
	}

	// Error handler for proxy failures
	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		log.Errorf("amp upstream proxy error for %s %s: %v", req.Method, req.URL.Path, err)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusBadGateway)
		_, _ = rw.Write([]byte(`{"error":"amp_upstream_proxy_error","message":"Failed to reach Amp upstream"}`))
	}

	return proxy, nil
}

// isStreamingResponse detects if the response is streaming (SSE only)
// Note: We only treat text/event-stream as streaming. Chunked transfer encoding
// is a transport-level detail and doesn't mean we can't decompress the full response.
// Many JSON APIs use chunked encoding for normal responses.
func isStreamingResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")

	// Only Server-Sent Events are true streaming responses
	if strings.Contains(contentType, "text/event-stream") {
		return true
	}

	return false
}

// proxyHandler converts httputil.ReverseProxy to gin.HandlerFunc
func proxyHandler(proxy *httputil.ReverseProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

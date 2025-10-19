package amp

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
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

		// Peek at the first few bytes to detect gzip magic bytes
		// without buffering the entire body
		var buf bytes.Buffer
		peeker := io.TeeReader(resp.Body, &buf)
		header := make([]byte, 2)
		n, err := io.ReadFull(peeker, header)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			// Restore original body on read error
			resp.Body = io.NopCloser(io.MultiReader(&buf, resp.Body))
			return nil
		}

		// Check for gzip magic bytes (0x1f 0x8b)
		if n >= 2 && header[0] == 0x1f && header[1] == 0x8b {
			// It's gzip - decompress the entire response
			// Read the rest from peeker (which includes the header we already read)
			fullBody, err := io.ReadAll(peeker)
			if err != nil {
				resp.Body = io.NopCloser(io.MultiReader(&buf, resp.Body))
				return nil
			}

			// Add the buffer content (header) to fullBody
			gzippedData := append(buf.Bytes(), fullBody...)

			// Decompress
			gzipReader, err := gzip.NewReader(bytes.NewReader(gzippedData))
			if err != nil {
				log.Warnf("amp proxy: gzip header detected but decompress failed: %v", err)
				// Return original body
				resp.Body = io.NopCloser(bytes.NewReader(gzippedData))
				return nil
			}

			decompressed, err := io.ReadAll(gzipReader)
			_ = gzipReader.Close()
			if err != nil {
				log.Warnf("amp proxy: gzip decompress error: %v", err)
				resp.Body = io.NopCloser(bytes.NewReader(gzippedData))
				return nil
			}

			// Replace body with decompressed content
			resp.Body = io.NopCloser(bytes.NewReader(decompressed))
			resp.ContentLength = int64(len(decompressed))
			resp.Header.Del("Content-Encoding") // Ensure it's clear this is not compressed
			log.Debugf("amp proxy: decompressed gzip response (%d -> %d bytes)", len(gzippedData), len(decompressed))
		} else {
			// Not gzip - restore original body with peeked bytes
			resp.Body = io.NopCloser(io.MultiReader(&buf, resp.Body))
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

// isStreamingResponse detects if the response is streaming (SSE or chunked)
func isStreamingResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	
	// Check for Server-Sent Events
	if strings.Contains(contentType, "text/event-stream") {
		return true
	}

	// Check for chunked transfer encoding
	if strings.Contains(strings.ToLower(resp.Header.Get("Transfer-Encoding")), "chunked") {
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

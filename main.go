package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

// Flags
var (
	sslMode       string
	sslCrt        string
	sslKey        string
	proxyAddress  string
	fixedUpstream string
	httpPort      int
	httpsPort     int
)

type ProxyHandler struct {
	httpClient *http.Client
	proxyURL   *url.URL // Parsed proxy URL for manual dialing
}

func main() {
	// Parse flags
	flag.StringVar(&sslMode, "ssl", "none", "SSL mode: generate, none, on")
	flag.StringVar(&sslCrt, "ssl-crt", "", "Path to SSL certificate (required if ssl=on)")
	flag.StringVar(&sslKey, "ssl-key", "", "Path to SSL key (required if ssl=on)")
	flag.StringVar(&proxyAddress, "proxy", "", "Upstream proxy URL (e.g., https://localhost:7897)")
	flag.StringVar(&fixedUpstream, "fixed-upstream", "", "Fixed upstream address (e.g., https://192.168.1.1) to forward requests to, while preserving the original Host header.")
	flag.IntVar(&httpPort, "http-port", 80, "HTTP listen port")
	flag.IntVar(&httpsPort, "https-port", 443, "HTTPS listen port")
	flag.Parse()

	// Validate flags
	if sslMode == "on" {
		if sslCrt == "" || sslKey == "" {
			log.Fatal("Error: --ssl-crt and --ssl-key are required when --ssl=on")
		}
	}

	// Parse proxy URL if provided
	var parsedProxyURL *url.URL
	var err error
	if proxyAddress != "" {
		parsedProxyURL, err = url.Parse(proxyAddress)
		if err != nil {
			log.Fatalf("Invalid proxy URL: %v", err)
		}
	}

	// Create Proxy Handler
	handler, err := NewProxyHandler(parsedProxyURL)
	if err != nil {
		log.Fatalf("Failed to create proxy handler: %v", err)
	}

	// Start Servers
	errChan := make(chan error)

	// HTTP Server
	go func() {
		addr := fmt.Sprintf(":%d", httpPort)
		log.Printf("Starting HTTP server on %s", addr)
		if err := http.ListenAndServe(addr, handler); err != nil {
			errChan <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	// HTTPS Server
	if sslMode != "none" {
		go func() {
			addr := fmt.Sprintf(":%d", httpsPort)
			log.Printf("Starting HTTPS server on %s", addr)

			var tlsConfig *tls.Config
			var err error

			if sslMode == "generate" {
				log.Println("Generating self-signed certificate for *.reserver.proxy")
				tlsConfig, err = generateTLSConfig()
				if err != nil {
					errChan <- fmt.Errorf("Failed to generate certificate: %v", err)
					return
				}
			} else { // sslMode == "on"
				cert, err := tls.LoadX509KeyPair(sslCrt, sslKey)
				if err != nil {
					errChan <- fmt.Errorf("Failed to load certificate: %v", err)
					return
				}
			tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
			}

			server := &http.Server{
				Addr:      addr,
				Handler:   handler,
				TLSConfig: tlsConfig,
			}
			if err := server.ListenAndServeTLS("", ""); err != nil {
				errChan <- fmt.Errorf("HTTPS server error: %v", err)
			}
		}()
	}

	if proxyAddress != "" {
		log.Printf("Upstream Proxy: %s", proxyAddress)
	}
	if fixedUpstream != "" {
		log.Printf("Fixed Upstream: %s (Host header will be preserved)", fixedUpstream)
	}

	// Wait for error (or block forever if no error)
	log.Fatal(<-errChan)
}

func NewProxyHandler(proxyURL *url.URL) (*ProxyHandler, error) {
	client, err := createHTTPClient(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("create HTTP client failed: %v", err)
	}

	return &ProxyHandler{
		httpClient: client,
		proxyURL:   proxyURL,
	}, nil
}

// createHTTPClient creates an HTTP client with proxy support
func createHTTPClient(proxyURL *url.URL) (*http.Client, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // Skip verify for upstream/proxy
	}

	transport := &http.Transport{
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	if proxyURL != nil {
		transport.Proxy = http.ProxyURL(proxyURL)
	} else {
		// Fallback to environment variables
		transport.Proxy = http.ProxyFromEnvironment
	}

	// Check SOCKS proxy from env if no specific proxy URL provided?
	if proxyURL == nil {
		socksProxy := os.Getenv("SOCKS_PROXY")
		if socksProxy != "" {
			dialer, err := createSOCKSDialer(socksProxy)
			if err != nil {
				return nil, fmt.Errorf("create SOCKS dialer failed: %v", err)
			}
			transport.DialContext = dialer.DialContext
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return client, nil
}

// generateTLSConfig generates a self-signed cert
func generateTLSConfig() (*tls.Config, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: "*.reserver.proxy",
		},
		DNSNames: []string{"*.reserver.proxy", "reserver.proxy"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}

// createSOCKSDialer creates a SOCKS5 dialer
func createSOCKSDialer(socksProxy string) (*socksDialer, error) {
	proxyURL, err := url.Parse(socksProxy)
	if err != nil {
		return nil, fmt.Errorf("parse SOCKS proxy failed: %v", err)
	}

	var auth *proxy.Auth
	if proxyURL.User != nil {
		password, _ := proxyURL.User.Password()
		auth = &proxy.Auth{
			User:     proxyURL.User.Username(),
			Password: password,
		}
	}

	dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("create SOCKS5 proxy failed: %v", err)
	}

	return &socksDialer{dialer: dialer}, nil
}

type socksDialer struct {
	dialer proxy.Dialer
}

func (d *socksDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return d.dialer.Dial(network, addr)
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("PANIC: %v", rec)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	log.Printf("===Request Start=== %s %s (Host: %s)", r.Method, r.URL.Path, r.Host)

	if isWebSocketRequest(r) {
		h.handleWebSocket(w, r)
		return
	}

	originalHost := r.Host
	if originalHost == "" {
		http.Error(w, "Host header is missing", http.StatusBadRequest)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if strings.HasSuffix(originalHost, ":443") {
		scheme = "https"
	}

	// Default target calculation
	targetHost := originalHost
	targetScheme := scheme

	// Override if fixedUpstream is set
	if fixedUpstream != "" {
		u, err := url.Parse(fixedUpstream)
		if err == nil {
			if u.Scheme != "" {
				targetScheme = u.Scheme
			}
			targetHost = u.Host
		} else {
			// Fallback: assume user provided host:port or just host. Use original scheme.
			targetHost = fixedUpstream
		}
	}

	targetURL := &url.URL{
		Scheme:   targetScheme,
		Host:     targetHost,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		log.Printf("Create proxy request failed: %v", err)
		http.Error(w, "Create proxy request failed", http.StatusInternalServerError)
		return
	}

	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Explicitly set the Host header to the original host
	proxyReq.Host = originalHost

	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		if prior := proxyReq.Header.Get("X-Forwarded-For"); prior != "" {
			clientIP = prior + ", " + clientIP
		}
		proxyReq.Header.Set("X-Forwarded-For", clientIP)
	}
	proxyReq.Header.Set("X-Forwarded-Proto", scheme)
	if r.Host != "" {
		proxyReq.Header.Set("X-Forwarded-Host", r.Host)
	}

	log.Printf("Forwarding to: %s (Host: %s)", targetURL.String(), proxyReq.Host)

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		log.Printf("Proxy request failed: %v", err)
		http.Error(w, "Proxy request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	written, err := io.Copy(w, resp.Body)
	if err != nil && !strings.Contains(err.Error(), "write: broken pipe") {
		log.Printf("Copy response body failed: %v", err)
		return
	}

	log.Printf("Request finished: %s %s -> %d (%d bytes)", r.Method, r.URL.Path, resp.StatusCode, written)
}

func isWebSocketRequest(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func (h *ProxyHandler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Printf("WebSocket Request: %s (Host: %s)", r.URL.Path, r.Host)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Printf("WebSocket Error: Hijacking not supported")
		http.Error(w, "WebSocket not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Printf("WebSocket Hijack failed: %v", err)
		http.Error(w, "Hijack failed", http.StatusInternalServerError)
		return
	}
	defer clientConn.Close()

	originalHost := r.Host
	if originalHost == "" {
		log.Printf("WebSocket missing Host header")
		clientConn.Write([]byte("HTTP/1.1 400 Bad Request\r\n\r\n"))
		return
	}

	scheme := "ws"
	if r.TLS != nil {
		scheme = "wss"
	}
	if strings.HasSuffix(originalHost, ":443") {
		scheme = "wss"
	}

	// Calculate upstream connection target
	connectHost := originalHost
	sniHost := originalHost // SNI usually matches the Host header

	if fixedUpstream != "" {
		u, err := url.Parse(fixedUpstream)
		if err == nil {
			if u.Scheme == "https" || u.Scheme == "wss" {
				scheme = "wss"
			} else if u.Scheme == "http" || u.Scheme == "ws" {
				scheme = "ws"
			}
			connectHost = u.Host
		} else {
			connectHost = fixedUpstream
		}
	}

	upstreamConn, err := h.dialUpstream(scheme, connectHost, sniHost)
	if err != nil {
		log.Printf("WebSocket dial upstream failed: %v", err)
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}
	defer upstreamConn.Close()

	upgradeReq := fmt.Sprintf("%s %s HTTP/1.1\r\n", r.Method, r.RequestURI)
	for key, values := range r.Header {
		if strings.ToLower(key) == "host" {
			continue
		}
		for _, value := range values {
			upgradeReq += fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}

	// Host header sent to upstream MUST be the original host
	if originalHost != "" {
		upgradeReq += fmt.Sprintf("Host: %s\r\n", originalHost)
	}
	upgradeReq += "\r\n"

	log.Printf("Sending WebSocket upgrade to upstream: Host=%s (Physically connecting to: %s)", originalHost, connectHost)

	if _, err := upstreamConn.Write([]byte(upgradeReq)); err != nil {
		log.Printf("WebSocket send upgrade failed: %v", err)
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}

	responseBuf := make([]byte, 4096)
	n, err := upstreamConn.Read(responseBuf)
	if err != nil {
		log.Printf("Read WebSocket upgrade response failed: %v", err)
		clientConn.Write([]byte("HTTP/1.1 502 Bad Gateway\r\n\r\n"))
		return
	}

	if _, err := clientConn.Write(responseBuf[:n]); err != nil {
		log.Printf("Forward WebSocket upgrade response failed: %v", err)
		return
	}

	responseStr := string(responseBuf[:n])
	if !strings.Contains(responseStr, "101") {
		log.Printf("WebSocket upgrade failed, upstream response: %s", strings.Split(responseStr, "\r\n")[0])
		return
	}

	log.Printf("WebSocket connected, bridging data")
	errChan := make(chan error, 2)

	go func() {
		_, err := io.Copy(upstreamConn, clientConn)
		errChan <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, upstreamConn)
		errChan <- err
	}()

	err = <-errChan
	if err != nil && !isConnectionClosedError(err) {
		log.Printf("WebSocket bridge error: %v", err)
	} else {
		log.Printf("WebSocket closed")
	}
}

func (h *ProxyHandler) dialUpstream(scheme, connectHost, sniHost string) (net.Conn, error) {
	// 1. Use Configured Proxy (CLI)
	if h.proxyURL != nil {
		conn, err := h.dialThroughHTTPProxy(h.proxyURL, connectHost)
		if err != nil {
			return nil, err
		}
		return h.wrapTLSIfNeeded(conn, scheme, connectHost, sniHost)
	}

	// 2. Use SOCKS Proxy (Env)
	socksProxy := os.Getenv("SOCKS_PROXY")
	if socksProxy != "" {
		dialer, err := createSOCKSDialer(socksProxy)
		if err != nil {
			return nil, fmt.Errorf("create SOCKS dialer failed: %v", err)
		}
		conn, err := dialer.dialer.Dial("tcp", connectHost)
		if err != nil {
			return nil, err
		}
		return h.wrapTLSIfNeeded(conn, scheme, connectHost, sniHost)
	}

	// 3. Use HTTP/HTTPS Proxy (Env)
	httpProxy := os.Getenv("HTTP_PROXY")
	httpsProxy := os.Getenv("HTTPS_PROXY")

	if httpProxy != "" || httpsProxy != "" {
		proxyURLStr := httpProxy
		if scheme == "wss" && httpsProxy != "" {
			proxyURLStr = httpsProxy
		}
		if proxyURLStr != "" {
			proxyURL, err := url.Parse(proxyURLStr)
			if err == nil {
				conn, err := h.dialThroughHTTPProxy(proxyURL, connectHost)
				if err != nil {
					return nil, err
				}
				return h.wrapTLSIfNeeded(conn, scheme, connectHost, sniHost)
			}
		}
	}

	// 4. Direct
	conn, err := net.DialTimeout("tcp", connectHost, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return h.wrapTLSIfNeeded(conn, scheme, connectHost, sniHost)
}

func (h *ProxyHandler) wrapTLSIfNeeded(conn net.Conn, scheme, connectHost, sniHost string) (net.Conn, error) {
	if scheme == "wss" || scheme == "https" {
		// Use sniHost for ServerName to ensure correct certificate from upstream
		tlsConn := tls.Client(conn, &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         strings.Split(sniHost, ":")[0],
		})
		if err := tlsConn.Handshake(); err != nil {
			conn.Close()
			return nil, fmt.Errorf("TLS handshake failed: %v", err)
		}
		return tlsConn, nil
	}
	return conn, nil
}

func (h *ProxyHandler) dialThroughHTTPProxy(proxyURL *url.URL, targetAddr string) (net.Conn, error) {
	proxyConn, err := net.DialTimeout("tcp", proxyURL.Host, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy failed: %v", err)
	}

	connectReq := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", targetAddr, targetAddr)
	if _, err := proxyConn.Write([]byte(connectReq)); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("send CONNECT failed: %v", err)
	}

	buf := make([]byte, 4096)
	n, err := proxyConn.Read(buf)
	if err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("read CONNECT response failed: %v", err)
	}

	response := string(buf[:n])
	if !strings.Contains(response, "200") {
		proxyConn.Close()
		return nil, fmt.Errorf("CONNECT failed: %s", strings.Split(response, "\r\n")[0])
	}

	return proxyConn, nil
}

func isConnectionClosedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "use of closed network connection") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "connection reset") ||
		err == io.EOF
}
package client

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/EduGoGroup/edugo-shared/auth"
)

// TokenInfo contains the result of a token validation.
type TokenInfo struct {
	Valid         bool              `json:"valid"`
	UserID        string            `json:"user_id,omitempty"`
	Email         string            `json:"email,omitempty"`
	ExpiresAt     time.Time         `json:"expires_at,omitempty"`
	Error         string            `json:"error,omitempty"`
	ActiveContext *auth.UserContext `json:"active_context,omitempty"`
}

// AuthClientConfig configures the auth client.
type AuthClientConfig struct {
	JWTSecret       string
	JWTIssuer       string
	BaseURL         string
	Timeout         time.Duration
	RemoteEnabled   bool
	FallbackEnabled bool
	CacheTTL        time.Duration
	CacheEnabled    bool
}

// AuthClient validates JWT tokens locally with optional remote fallback.
type AuthClient struct {
	jwtManager *auth.JWTManager
	baseURL    string
	httpClient *http.Client
	cache      *tokenCache
	config     AuthClientConfig
}

// NewAuthClient creates a new auth client instance.
func NewAuthClient(cfg AuthClientConfig) *AuthClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 60 * time.Second
	}
	if cfg.JWTIssuer == "" {
		cfg.JWTIssuer = "edugo-central"
	}

	var jwtMgr *auth.JWTManager
	if cfg.JWTSecret != "" {
		jwtMgr = auth.NewJWTManager(cfg.JWTSecret, cfg.JWTIssuer)
	}

	return &AuthClient{
		jwtManager: jwtMgr,
		baseURL:    cfg.BaseURL,
		httpClient: &http.Client{Timeout: cfg.Timeout},
		cache:      newTokenCache(cfg.CacheTTL),
		config:     cfg,
	}
}

// ValidateToken validates a JWT token using local validation first, with optional remote fallback.
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	cacheKey := hashToken(token)
	if c.config.CacheEnabled {
		if cached, found := c.cache.Get(cacheKey); found {
			return cached, nil
		}
	}

	// Local validation (preferred)
	if c.jwtManager != nil {
		info, err := c.validateLocally(token)
		if err == nil && info.Valid {
			if c.config.CacheEnabled {
				c.cache.Set(cacheKey, info)
			}
			return info, nil
		}

		if !c.config.FallbackEnabled || !c.config.RemoteEnabled {
			if info != nil {
				return info, nil
			}
			return &TokenInfo{Valid: false, Error: fmt.Sprintf("local validation failed: %v", err)}, nil
		}
	}

	// Remote fallback
	if c.config.RemoteEnabled && c.baseURL != "" {
		info, err := c.validateRemotely(ctx, token)
		if err != nil {
			return &TokenInfo{Valid: false, Error: fmt.Sprintf("remote validation failed: %v", err)}, nil
		}
		if c.config.CacheEnabled && info.Valid {
			c.cache.Set(cacheKey, info)
		}
		return info, nil
	}

	return &TokenInfo{Valid: false, Error: "no validation method available"}, nil
}

func (c *AuthClient) validateLocally(token string) (*TokenInfo, error) {
	claims, err := c.jwtManager.ValidateToken(token)
	if err != nil {
		return &TokenInfo{Valid: false, Error: err.Error()}, err
	}

	var expiresAt time.Time
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}

	return &TokenInfo{
		Valid:         true,
		UserID:        claims.UserID,
		Email:         claims.Email,
		ExpiresAt:     expiresAt,
		ActiveContext: claims.ActiveContext,
	}, nil
}

func (c *AuthClient) validateRemotely(ctx context.Context, token string) (*TokenInfo, error) {
	url := c.baseURL + "/v1/auth/verify"
	bodyBytes, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling auth service: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("auth service error: status %d", resp.StatusCode)
	}

	var info TokenInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &info, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// tokenCache is a simple TTL cache for validated tokens.
type tokenCache struct {
	entries map[string]*cacheEntry
	ttl     time.Duration
	mu      sync.RWMutex
}

type cacheEntry struct {
	info      *TokenInfo
	expiresAt time.Time
}

func newTokenCache(ttl time.Duration) *tokenCache {
	c := &tokenCache{
		entries: make(map[string]*cacheEntry),
		ttl:     ttl,
	}
	go c.cleanupLoop()
	return c
}

func (c *tokenCache) Get(key string) (*TokenInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.info, true
}

func (c *tokenCache) Set(key string, info *TokenInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = &cacheEntry{info: info, expiresAt: time.Now().Add(c.ttl)}
}

func (c *tokenCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.entries {
			if now.After(e.expiresAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}

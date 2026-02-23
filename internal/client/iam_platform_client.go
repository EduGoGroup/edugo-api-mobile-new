package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IAMPlatformConfig configures the IAM platform client.
type IAMPlatformConfig struct {
	BaseURL string
	Timeout time.Duration
}

// IAMPlatformClient is an HTTP client for the IAM platform API.
type IAMPlatformClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewIAMPlatformClient creates a new IAM platform client.
func NewIAMPlatformClient(cfg IAMPlatformConfig) *IAMPlatformClient {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	return &IAMPlatformClient{
		baseURL:    cfg.BaseURL,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// CombinedScreenResponse represents a resolved screen from iam-platform.
type CombinedScreenResponse struct {
	ScreenID     string          `json:"screen_id"`
	ScreenKey    string          `json:"screen_key"`
	ScreenName   string          `json:"screen_name"`
	Pattern      string          `json:"pattern"`
	Version      int             `json:"version"`
	Template     json.RawMessage `json:"template"`
	SlotData     json.RawMessage `json:"slot_data"`
	Actions      json.RawMessage `json:"actions"`
	DataEndpoint string          `json:"data_endpoint,omitempty"`
	DataConfig   json.RawMessage `json:"data_config,omitempty"`
	HandlerKey   *string         `json:"handler_key,omitempty"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// ResourceScreenResponse represents a resource-screen link from iam-platform.
type ResourceScreenResponse struct {
	ResourceID  string `json:"resource_id"`
	ResourceKey string `json:"resource_key"`
	ScreenKey   string `json:"screen_key"`
	ScreenType  string `json:"screen_type"`
	IsDefault   bool   `json:"is_default"`
}

// MenuItemResponse represents a menu item from iam-platform.
type MenuItemResponse struct {
	Key         string             `json:"key"`
	DisplayName string             `json:"display_name"`
	Icon        string             `json:"icon,omitempty"`
	Scope       string             `json:"scope"`
	SortOrder   int                `json:"sort_order"`
	Permissions []string           `json:"permissions,omitempty"`
	Screens     map[string]string  `json:"screens,omitempty"`
	Children    []MenuItemResponse `json:"children"`
}

// MenuResponse wraps menu items.
type MenuResponse struct {
	Items []MenuItemResponse `json:"items"`
}

// ResolveScreenByKey gets a combined screen config from iam-platform.
func (c *IAMPlatformClient) ResolveScreenByKey(ctx context.Context, key, authToken string) (*CombinedScreenResponse, error) {
	url := c.baseURL + "/v1/screen-config/resolve/key/" + key
	body, err := c.doGet(ctx, url, authToken)
	if err != nil {
		return nil, err
	}
	var result CombinedScreenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing screen response: %w", err)
	}
	return &result, nil
}

// GetScreensByResource gets screens for a resource from iam-platform.
func (c *IAMPlatformClient) GetScreensByResource(ctx context.Context, resourceID, authToken string) ([]*ResourceScreenResponse, error) {
	url := c.baseURL + "/v1/screen-config/resource-screens/" + resourceID
	body, err := c.doGet(ctx, url, authToken)
	if err != nil {
		return nil, err
	}
	var result []*ResourceScreenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing resource screens response: %w", err)
	}
	return result, nil
}

// GetMenu gets the menu filtered by permissions from iam-platform.
func (c *IAMPlatformClient) GetMenu(ctx context.Context, authToken string) (*MenuResponse, error) {
	url := c.baseURL + "/v1/menu"
	body, err := c.doGet(ctx, url, authToken)
	if err != nil {
		return nil, err
	}
	var result MenuResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing menu response: %w", err)
	}
	return &result, nil
}

func (c *IAMPlatformClient) doGet(ctx context.Context, url, authToken string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calling iam-platform: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("iam-platform error: status %d, body: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

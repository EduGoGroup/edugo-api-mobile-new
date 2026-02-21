package dto

import (
	"encoding/json"
)

// ScreenResponse is the API response for a composed screen.
type ScreenResponse struct {
	ScreenKey   string          `json:"screen_key"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Pattern     string          `json:"pattern"`
	Definition  json.RawMessage `json:"definition"`
	SlotData    json.RawMessage `json:"slot_data"`
	Actions     json.RawMessage `json:"actions,omitempty"`
	IsActive    bool            `json:"is_active"`
}

// NavigationNode represents a single node in the navigation tree.
type NavigationNode struct {
	Key         string            `json:"key"`
	DisplayName string            `json:"display_name"`
	Icon        *string           `json:"icon,omitempty"`
	SortOrder   int               `json:"sort_order"`
	Screens     []NavigationScreen `json:"screens,omitempty"`
	Children    []NavigationNode   `json:"children,omitempty"`
}

// NavigationScreen is a screen reference inside a navigation node.
type NavigationScreen struct {
	ScreenKey  string `json:"screen_key"`
	ScreenType string `json:"screen_type"`
	IsDefault  bool   `json:"is_default"`
	SortOrder  int    `json:"sort_order"`
}

// SavePreferencesRequest is the payload for saving screen preferences.
type SavePreferencesRequest struct {
	Preferences json.RawMessage `json:"preferences" binding:"required"`
}

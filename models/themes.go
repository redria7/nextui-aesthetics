package models

import (
	//"go.uber.org/zap/zapcore"
)

type Theme struct {
	ThemeName		string
	ThemePath		string
	PreviewFound	bool
	ContainsTheme	bool
	IsHidden		bool
}

type Component struct {
	ComponentName 	string
	IsSupported		bool
	ComponentPaths	[]string
	ComponentType	ComponentTypeDetails
	// ComponentCount	int	// future improvement?
}

type ComponentTypeDetails struct {
	ComponentType			string
	ComponentHomeDirectory	string
	ContainsMetaFiles		bool
	DuplicateType			bool
}

type ComponentOptionSelections struct{
	OptionAll		bool
	OptionActive	bool
	OptionInactive	bool
	OptionClear		bool
	OptionPreserve	bool
	OptionConfirm	bool
}

// CatalogData represents the structure of the catalog.json file
type CatalogData struct {
	Themes		map[string]CatalogItemInfo            `json:"themes"`		// .Themes[ThemeName] = catalogiteminfo					theme type is Theme,	theme name is key
	Components	map[string]map[string]CatalogItemInfo `json:"components"`	// .Components[ThemeType][ThemeName] = catalogiteminfo	theme type is key,		theme name is key
}

// CatalogItemInfo represents an item in the catalog
type CatalogItemInfo struct {
	PreviewPath	string `json:"preview_path"`
	Author		string `json:"author"`
	Description	string `json:"description"`
	URL			string `json:"URL"`
	LastUpdated	string `json:"last_updated"`
}

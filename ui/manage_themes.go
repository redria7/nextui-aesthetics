package ui

import (
	"sort"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"qlova.tech/sum"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
)

const (
	// DownloadThemesDisplayName	= "Download Themes"
	// ManageThemesDisplayName		= "Manage Themes"
	// DecorationsDisplayName 		= "Set Wallpapers & Icons"
)

type ManageThemes struct{}

func InitManageThemes() ManageThemes {
	return ManageThemes{}
}

func (mt ManageThemes) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageThemes
}

func (mt ManageThemes) Draw() (interface{}, int, error) {
	title := "Manage Themes"

	// Add items to menu
	currentThemes := utils.GetDownloadedThemes()
	themeKeys := make([]string, len(currentThemes))
	keyIndex := 0
	for key, _ := range currentThemes {
		themeKeys[keyIndex] = key
		keyIndex++
	}
	sort.Strings(themeKeys)
	var menuItems []gaba.MenuItem
	for _, key := range themeKeys {
		theme := currentThemes[key]
		if theme.ContainsTheme {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     theme.ThemeName,
				Selected: false,
				Focused:  false,
				Metadata: theme,
				ImageFilename: utils.GetPreviewPath(theme.ThemeName),
			})
		}
	}

	// Set options
	options := gaba.DefaultListOptions(title, menuItems)
	options.SmallTitle = true
	options.EnableAction = true
	options.EmptyMessage = "No themes to manage! Save or download some!"
	options.EnableImages = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "X", HelpText: "Delete"},
		{ButtonName: "A", HelpText: "Select"},
	}

	// Wait for results
	selection, err := gaba.List(options)

	// Handle error
	if err != nil {
		return nil, utils.ExitCodeError, err
	}

	// Process successful results
	if selection.IsSome() && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		exit_code := utils.ExitCodeAction
		if !selection.Unwrap().ActionTriggered {
			exit_code = utils.ExitCodeSelect
		}
		return selection.Unwrap().SelectedItem.Metadata.(models.Theme), exit_code, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

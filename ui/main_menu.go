package ui

import (
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"qlova.tech/sum"
	"nextui-aesthetics/internal/i18n"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
)

const (
	DownloadThemesDisplayName		= "Download Themes"
	ManageThemesDisplayName			= "Manage Available Themes"
	ManageCurrentThemeDisplayName	= "Manage Current Theme"
	DecorationsDisplayName 			= "Set Wallpapers & Icons"
)

type MainMenu struct{}

func InitMainMenu() MainMenu {
	return MainMenu{}
}

func (m MainMenu) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.MainMenu
}

func (m MainMenu) Draw() (interface{}, int, error) {
	title := "Aesthetics"

	// Add items to menu (Text is translated, Metadata stays in English so
	// the switch-case dispatch in app/aesthetics.go keeps working).
	var menuItems []gaba.MenuItem
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     i18n.T("ae.menu.download_themes"),
		Selected: false,
		Focused:  false,
		Metadata: DownloadThemesDisplayName,
	})
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     i18n.T("ae.menu.manage_themes"),
		Selected: false,
		Focused:  false,
		Metadata: ManageThemesDisplayName,
	})
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     i18n.T("ae.menu.manage_current"),
		Selected: false,
		Focused:  false,
		Metadata: ManageCurrentThemeDisplayName,
	})
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     i18n.T("ae.menu.decorations"),
		Selected: false,
		Focused:  false,
		Metadata: DecorationsDisplayName,
	})

	// Set options
	options := gaba.DefaultListOptions(title, menuItems)
	options.EnableAction = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: i18n.T("ae.btn.quit")},
		{ButtonName: "X", HelpText: i18n.T("ae.btn.settings")},
		{ButtonName: "A", HelpText: i18n.T("ae.btn.select")},
	}

	// Wait for results
	selection, err := gaba.List(options)

	// Handle error
	if err != nil {
		return nil, utils.ExitCodeError, err
	}

	// Process successful results
	if selection.IsSome() && selection.Unwrap().ActionTriggered {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return nil, utils.ExitCodeAction, nil
	} else if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return selection.Unwrap().SelectedItem.Metadata.(string), utils.ExitCodeSelect, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

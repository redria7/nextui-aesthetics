package ui

import (
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"qlova.tech/sum"
	"nextui-aesthetics/models"
	"nextui-aesthetics/internal/i18n"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
)

const (
	DeleteDisplayName	= "Delete Theme"
	RenameDisplayName	= "Rename Theme"
)

type ManageThemeOptions struct{
	Theme 	models.Theme
}

func InitManageThemeOptions(theme models.Theme) ManageThemeOptions {
	return ManageThemeOptions{
		Theme:		theme,
	}
}

func (mto ManageThemeOptions) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageThemeOptions
}

func (mto ManageThemeOptions) Draw() (interface{}, int, error) {
	title := mto.Theme.ThemeName + " Options"

	// Add items to menu
	var menuItems []gaba.MenuItem
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     i18n.T(RenameDisplayName),
		Selected: false,
		Focused:  false,
		Metadata: RenameDisplayName,
	})
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     i18n.T(DeleteDisplayName),
		Selected: false,
		Focused:  false,
		Metadata: DeleteDisplayName,
	})

	// Set options
	options := gaba.DefaultListOptions(title, menuItems)
	options.SmallTitle = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: i18n.T("ae.btn.back")},
		{ButtonName: "A", HelpText: i18n.T("ae.btn.select")},
	}

	// Wait for results
	selection, err := gaba.List(options)

	// Handle error
	if err != nil {
		return nil, utils.ExitCodeError, err
	}

	// Process successful results
	if selection.IsSome() && !selection.Unwrap().ActionTriggered && selection.Unwrap().SelectedIndex != -1 {
		state.UpdateCurrentMenuPosition(selection.Unwrap().SelectedIndex, selection.Unwrap().VisiblePosition)
		return selection.Unwrap().SelectedItem.Metadata.(string), utils.ExitCodeSelect, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

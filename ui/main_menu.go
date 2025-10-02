package ui

import (
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"qlova.tech/sum"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
)

const (
	DecorationsDisplayName = "Library Decorations"
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

	// Add items to menu
	var menuItems []gaba.MenuItem
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     DecorationsDisplayName,
		Selected: false,
		Focused:  false,
		Metadata: DecorationsDisplayName,
	})

	// Set options
	options := gaba.DefaultListOptions(title, menuItems)

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
		{ButtonName: "A", HelpText: "Select"},
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
		return selection.Unwrap().SelectedItem.Metadata.(shared.RomDirectory), utils.ExitCodeSelect, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

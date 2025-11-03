package ui

import (
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

type ManageThemeComponents struct{
	Theme 	models.Theme
}

func InitManageThemeComponents(theme models.Theme) ManageThemeComponents {
	return ManageThemeComponents{
		Theme:	theme,
	}
}

func (mtc ManageThemeComponents) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageThemeComponents
}

func (mtc ManageThemeComponents) Draw() (interface{}, int, error) {
	// Set values depending on selected theme
	themeTitle := "Current"
	actionOption := "Clear"
	selectOption := "Save"
	selectHelp := "Save selected components into a theme"
	actionHelp := "Revert selected components to default settings"
	if mtc.Theme != (models.Theme{}) {
		themeTitle = mtc.Theme.ThemeName
		actionOption = "Delete"
		selectOption = "Apply"
		selectHelp = "Apply selected components to device"
		actionHelp = "Delete selected components from the theme"
	}
	title := "Manage " + themeTitle + " Components"

	// Add items to menu
	components := utils.GetThemeComponents(mtc.Theme)
	var menuItems []gaba.MenuItem
	for _, component := range components {
		if component.IsSupported {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     component.ComponentName,
				Selected: false,
				Focused:  false,
				Metadata: component,
			})
		}
	}

	// Set options
	options := gaba.DefaultListOptions(title, menuItems)
	options.SmallTitle = true
	options.EnableAction = true
	options.EmptyMessage = "No supported components!"
	// Multiselect fixed options
	options.EnableMultiSelect = true
	options.StartInMultiSelectMode = true
	options.MultiSelectButton = gaba.ButtonUnassigned

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Toggle"},
		{ButtonName: "X", HelpText: actionOption},
		{ButtonName: "Start", HelpText: selectOption},
	}
	
	// Set Help
	options.EnableHelp = true
	options.HelpTitle = "Component Management Controls"
	options.HelpText = []string{
		"• A: Toggle a component to include in the selected action",
		"• Start: " + selectHelp,
		"• X: " + actionHelp,
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
		var metaReturn []models.Component
		selections := selection.Unwrap().SelectedItems
		for _, selection := range selections {
			metaReturn = append(metaReturn, selection.Metadata.(models.Component))
		}
		return metaReturn, exit_code, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

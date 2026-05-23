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
	themeTitle := i18n.T("ae.label.current")
	actionOption := i18n.T("ae.btn.clear")
	selectOption := i18n.T("ae.btn.save")
	selectHelp := i18n.T("ae.help.save_components")
	actionHelp := i18n.T("ae.help.revert_components")
	if mtc.Theme != (models.Theme{}) {
		themeTitle = mtc.Theme.ThemeName
		actionOption = i18n.T("ae.btn.delete")
		selectOption = i18n.T("ae.btn.apply")
		selectHelp = i18n.T("ae.help.apply_components")
		actionHelp = i18n.T("ae.help.delete_components")
	}
	title := i18n.T("ae.title.manage_components_prefix") + " " + themeTitle + " " + i18n.T("ae.title.manage_components_suffix")

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
	options.EmptyMessage = i18n.T("ae.empty.no_components")
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
		{ButtonName: "B", HelpText: i18n.T("ae.btn.back")},
		{ButtonName: "A", HelpText: i18n.T("ae.btn.toggle")},
		{ButtonName: "X", HelpText: actionOption},
		{ButtonName: "Start", HelpText: selectOption},
	}

	// Set Help
	options.EnableHelp = true
	options.HelpTitle = i18n.T("ae.help.component_management")
	options.HelpText = []string{
		"• A: " + i18n.T("ae.help.toggle"),
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

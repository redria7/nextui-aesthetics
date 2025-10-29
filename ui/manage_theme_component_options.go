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
	SaveForAllName			= "Save components: for all content"
	SaveForContentOnlyName	= "Save components: for active content"
	ClearContentOnlyName	= "Revert components to default: for active content"
	ClearNonContentOnlyName	= "Revert components to default: for inactive content"
	ClearAllName			= "Revert components to default: for all content"
	ApplyActiveAndClearName	= "Clear components then apply: for active content"
	ApplyAllAndClearName	= "Clear components then apply: for all content"
	ApplyActiveOverwrite	= "Apply components without clearing: for active content"
	ApplyAllOverwrite		= "Apply components without clearing: for all content"
	ApplyActivePreserve		= "Apply only missing components: for active content"
	ApplyAllPreserve		= "Apply only missing components: for all content"
)

type ManageThemeComponentOptions struct{
	Theme 	models.Theme
	Components	[]models.Component
	ClearSelected	bool
}

func InitManageThemeComponentOptions(theme models.Theme, components []models.Component, clearSelected bool) ManageThemeComponentOptions {
	return ManageThemeComponentOptions{
		Theme:	theme,
		Components: components,
		ClearSelected: clearSelected,
	}
}

func (mtc ManageThemeComponentOptions) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.ManageThemeComponentOptions
}

func (mtco ManageThemeComponentOptions) Draw() (interface{}, int, error) {
	isCurrentTheme := true
	if mtco.Theme != (models.Theme{}) {
		isCurrentTheme = false
	}

	// Add items to menu
	var menuItems []gaba.MenuItem
	if isCurrentTheme {
		if mtco.ClearSelected {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     ClearContentOnlyName,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionClear: true,
					OptionActive: true,
				},
			})
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     ClearNonContentOnlyName,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionClear: true,
					OptionInactive: true,
				},
			})
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     ClearAllName,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionClear: true,
					OptionAll: true,
				},
			})
		} else {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     SaveForContentOnlyName,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionActive: true,
				},
			})
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     SaveForAllName,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionAll: true,
				},
			})
		}
	} else {
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ApplyActiveAndClearName,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionClear: true,
				OptionActive: true,
			},
		})
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ApplyAllAndClearName,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionClear: true,
				OptionAll: true,
			},
		})
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ApplyActiveOverwrite,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionActive: true,
			},
		})
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ApplyAllOverwrite,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionAll: true,
			},
		})
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ApplyActivePreserve,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionActive: true,
				OptionPreserve: true,
			},
		})
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ApplyAllPreserve,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionAll: true,
				OptionPreserve: true,
			},
		})
	}

	// Set options
	options := gaba.DefaultListOptions("Component Options", menuItems)
	options.SmallTitle = true
	options.EmptyMessage = "No supported components!"

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Select"},
	}
	
	// Set Help
	options.EnableHelp = true
	options.HelpTitle = "Component Management Options"
	options.HelpText = []string{
		"'Active' content is ROM directories that contain ROMs as some depth",
		"Reverting to default will clear applied theme images",
		"Clearing before applying will clear applied theme images for a clean application",
		"Applying without clearing will only overwrite current images found in the selected theme components",
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
		return selection.Unwrap().SelectedItem.Metadata.(models.ComponentOptionSelections), utils.ExitCodeSelect, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

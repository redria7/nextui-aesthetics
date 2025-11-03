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
	SaveForAllName			= "Save: All"
	SaveForContentOnlyName	= "Save: + Active Consoles"
	SaveAllConfirm			= "Save: All: With Confirmations"
	ClearContentOnlyName	= "Revert to Default | + Active Consoles"
	ClearNonContentOnlyName	= "Revert to Default | Empty Consoles Only"
	ClearAllName			= "Revert to Default | All"
	ClearAllConfirm			= "Revert to Default | All | With Confirmations"
	ApplyActiveAndClearName	= "Clear & Apply | + Active Consoles"
	ApplyAllAndClearName	= "Clear & Apply | All"
	ApplyActiveOverwrite	= "Apply | + Active Consoles"
	ApplyAllOverwrite		= "Apply | All"
	ApplyActivePreserve		= "Apply | Missing Only | + Active Consoles"
	ApplyAllPreserve		= "Apply | Missing Only | All"
	ApplyAllConfirm			= "Apply | All | With Confirmations"
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
				Text:     ClearAllConfirm,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionClear: true,
					OptionAll: true,
					OptionConfirm: true,
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
				Text:     ClearContentOnlyName,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionClear: true,
					OptionActive: true,
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
				Text:     SaveAllConfirm,
				Selected: false,
				Focused:  false,
				Metadata: models.ComponentOptionSelections{
					OptionAll: true,
					OptionConfirm: true,
				},
			})
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
			Text:     ApplyAllConfirm,
			Selected: false,
			Focused:  false,
			Metadata: models.ComponentOptionSelections{
				OptionAll: true,
				OptionConfirm: true,
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
		"• 'Active Consoles' have roms inside!",
		"• 'Empty Consoles' have no roms :(",
		"• 'Save' saves a new theme in:",
		"      .userdata/shared/Aesthetics/Themes",
		"• 'Revert to Default' returns component(s) to",
		"      NextUI default settings",
		"• 'Apply' applies the selected components",
		"• 'Clear & Apply' reverts to default before applying",
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

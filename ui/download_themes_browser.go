package ui

import (
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"qlova.tech/sum"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
)

const (
	RefreshCatalogName = "Refresh Available Themes"
	ShowHiddenThemesName = "Hidden Themes"
	ExitCodeSpecialResult	= 5
)

type DownloadThemesBrowser struct{
	ShowHiddenThemes	bool
}

func InitDownloadThemesBrowser(showHiddenThemes bool) DownloadThemesBrowser {
	return DownloadThemesBrowser{
		ShowHiddenThemes: showHiddenThemes,
	}
}

func (dtb DownloadThemesBrowser) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadThemesBrowser
}

func (dtb DownloadThemesBrowser) Draw() (interface{}, int, error) {
	// Collect lists of themes available from the catalog, sorted into downloaded, new, and not downloaded buckets
	themeCatalog := state.GetThemeCatalog()
	currentThemes := utils.GetDownloadedThemes()
	var newThemes []gaba.MenuItem
	var downloadedThemes []gaba.MenuItem
	var notDownloadedThemes []gaba.MenuItem
	var hiddenThemes []gaba.MenuItem
	for _, theme := range themeCatalog {
		themeStatus, exists := currentThemes[theme.ThemeName]
		if exists {
			if themeStatus.ContainsTheme {
				downloadedThemes = append(downloadedThemes, gaba.MenuItem{
					Text:     "(*) " + theme.ThemeName,
					Selected: false,
					Focused:  false,
					Metadata: theme,
					ImageFilename: utils.GetPreviewPath(theme.ThemeName),
				})
			} else if themeStatus.IsHidden {
				hiddenThemes = append(hiddenThemes, gaba.MenuItem{
					Text:     theme.ThemeName,
					Selected: false,
					Focused:  false,
					Metadata: theme,
					ImageFilename: utils.GetPreviewPath(theme.ThemeName),
				})
			}else if theme.IsNew {
				newThemes = append(newThemes, gaba.MenuItem{
					Text:     "(NEW!) " + theme.ThemeName,
					Selected: false,
					Focused:  false,
					Metadata: theme,
					ImageFilename: utils.GetPreviewPath(theme.ThemeName),
				})
			} else {
				notDownloadedThemes = append(notDownloadedThemes, gaba.MenuItem{
					Text:     theme.ThemeName,
					Selected: false,
					Focused:  false,
					Metadata: theme,
					ImageFilename: utils.GetPreviewPath(theme.ThemeName),
				})
			}
		} else {
			newThemes = append(newThemes, gaba.MenuItem{
				Text:     "(NEW!) " + theme.ThemeName,
				Selected: false,
				Focused:  false,
				Metadata: theme,
				ImageFilename: "",
			})
		}
	}

	// Present list of new, then not downloaded, then downloaded
	var menuItems []gaba.MenuItem
	if !dtb.ShowHiddenThemes {
		menuItems = append(menuItems, newThemes...)
		menuItems = append(menuItems, notDownloadedThemes...)
		menuItems = append(menuItems, downloadedThemes...)
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     RefreshCatalogName,
			Selected: false,
			Focused:  false,
			Metadata: RefreshCatalogName,
		})
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     ShowHiddenThemesName,
			Selected: false,
			Focused:  false,
			Metadata: ShowHiddenThemesName,

		})
	} else {
		menuItems = append(menuItems, hiddenThemes...)
	}
	

	// Set options
	title := "Downloadable Themes"
	options := gaba.DefaultListOptions(title, menuItems)
	options.EnableImages = true
	options.EnableAction = true
	options.SmallTitle = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	actionText := "Hide Theme"
	if dtb.ShowHiddenThemes {
		actionText = "Unhide Theme"
	}
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: "Details"},
		{ButtonName: "X", HelpText: actionText},
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
		metadata := selection.Unwrap().SelectedItem.Metadata
		if metadata == RefreshCatalogName || metadata == ShowHiddenThemesName {
			return metadata.(string), ExitCodeSpecialResult, nil
		}
		if !selection.Unwrap().ActionTriggered {
			exit_code = utils.ExitCodeSelect
		}
		return metadata.(models.ThemeSummary), exit_code, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

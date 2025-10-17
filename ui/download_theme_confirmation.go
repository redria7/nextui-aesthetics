package ui

import (
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	//"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"nextui-aesthetics/models"
	"nextui-aesthetics/utils"
	"qlova.tech/sum"
)

type DownloadThemeConfirmation struct {
	ShowHiddenThemes	bool
	Theme				models.ThemeSummary
}

func InitDownloadThemeConfirmation(showHiddenThemes bool, theme models.ThemeSummary) DownloadThemeConfirmation {
	return DownloadThemeConfirmation{
		ShowHiddenThemes: 	showHiddenThemes,
		Theme:      		theme,
	}
}

func (dtc DownloadThemeConfirmation) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DownloadThemeConfirmation
}

func (dtc DownloadThemeConfirmation) Draw() (selection interface{}, exitCode int, e error) {
	// Collect info about current theme
	title := dtc.Theme.ThemeName
	currentThemes := utils.GetDownloadedThemes()
	themeStatus, exists := currentThemes[dtc.Theme.ThemeName]
	downloaded := false
	if exists && themeStatus.ContainsTheme {
		downloaded = true
	}

	// Prep info sections
	var sections []gaba.Section
	sections = append(sections, gaba.NewDescriptionSection(
		"",
		dtc.Theme.Description,
	))
	sections = append(sections, gaba.NewImageSection(
		"",
		utils.GetPreviewPath(dtc.Theme.PreviewPath),
		int32(256),
		int32(256),
		gaba.TextAlignCenter,
	))
	sections = append(sections, gaba.NewInfoSection(
		"",
		[]gaba.MetadataItem{
			{Label: "File Type", 	Value: dtc.Theme.ThemeType},
			{Label: "Author", 		Value: dtc.Theme.Author},
			{Label: "Last Updated", Value: dtc.Theme.LastUpdated},
		},
	))

	// Set options
	options := gaba.DefaultInfoScreenOptions()
	options.Sections = sections
	options.ShowThemeBackground = false

	// Set footers
	confirmLabel := "Download"
	if downloaded {
		confirmLabel = "Re-Download"
	}
	footerItems := []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: confirmLabel},
	}

	// Wait for results
	sel, err := gaba.DetailScreen(title, options, footerItems)

	// Handle error
	if err != nil {
		return nil, utils.ExitCodeError, err
	}

	// Process successful results
	if sel.IsNone() {
		return nil, utils.ExitCodeCancel, nil
	}

	return nil, utils.ExitCodeSelect, nil
}

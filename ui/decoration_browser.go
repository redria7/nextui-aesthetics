package ui

import (
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"qlova.tech/sum"
)

const (
	DefaultDecorationBrowserIndex = -1
)

type DecorationBrowser struct {
	RomDirectoryList		[]shared.RomDirectory
	ListWallpaperSelected	bool
	DecorationType			string
	DecorationBrowserIndex	int
}

func InitDecorationBrowser(romDirectoryList []shared.RomDirectory, listWallpaperSelected bool, decorationType string, decorationbrowserIndex int) DecorationBrowser {
	return DecorationBrowser{
		RomDirectoryList:	romDirectoryList,
		ListWallpaperSelected:	listWallpaperSelected,
		DecorationType: decorationType,
		DecorationBrowserIndex: decorationbrowserIndex,
	}
}

func (db DecorationBrowser) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DecorationBrowser
}

func (db DecorationBrowser) Draw() (item interface{}, exitCode int, e error) {
	//logger := common.GetLoggerInstance()
	topLevel := false
	if db.DecorationBrowserIndex == DefaultDecorationBrowserIndex {
		topLevel = true
	}
	currentDirectory := db.RomDirectoryList[len(db.RomDirectoryList) - 1]

	// Add items to menu
	var menuItems []gaba.MenuItem
	aggregationType := state.GetAppState().Config.DecorationAggregationType
	if aggregationType == utils.AggregateByConsole {
		menuItems = db.genConsoleMenuItems()
	}
	if aggregationType == utils.AggregateByDirectory {
		menuItems = db.genDirectoryMenuItems()
	}

	// Set options
	title := currentDirectory.DisplayName
	options := gaba.DefaultListOptions(title, menuItems)
	options.SmallTitle = true
	options.EmptyMessage = "No Decorations Found"
	options.EnableAction = true
	options.EnableImages = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	selectText := "Apply"
	actionText := "Delete"
	if topLevel {
		selectText = "View Decorations"
		actionText = "Change Aggregation"
	}
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
		{ButtonName: "A", HelpText: selectText},
		{ButtonName: "X", HelpText: actionText},
	}

	// Set Help
	options.EnableHelp = true
	options.HelpTitle = "Decoration List Controls"
	helpA := "Open confirmation screen to apply the selected decoration"
	helpX := "Open confirmation screen to delete the selected decoration"
	if topLevel {
		helpA = "Open selected aggregation to view available decorations"
		helpX = "Change aggregation style to view console or directory groupings"
	}
	options.HelpText = []string{
		"• A: " + helpA,
		"• X: " + helpX,
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
		if !selection.Unwrap().ActionTriggered {
			exit_code = utils.ExitCodeSelect
		}
		if topLevel {
			return metadata.(int), exit_code, nil
		}
		return metadata.(models.Decoration), exit_code, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

func (db DecorationBrowser) genConsoleMenuItems() ([]gaba.MenuItem) {
	var menuItems []gaba.MenuItem
	decorationAggregation, _ := state.GetDecorationAggregation()
	topLevel := false
	if db.DecorationBrowserIndex == DefaultDecorationBrowserIndex {
		topLevel = true
	}
	if topLevel {
		for index, aggregate := range decorationAggregation {
			menuItems = append(menuItems, gaba.MenuItem{
				Text: aggregate.ConsoleName,
				Selected: false,
				Focused:  false,
				Metadata: index,
			})
		}
	} else {
		_, currentPath, parentPath := utils.GetCurrentDecorationDetails(db.RomDirectoryList)
		currentDirectory := db.RomDirectoryList[len(db.RomDirectoryList) - 1]
		for _, decoration := range decorationAggregation[db.DecorationBrowserIndex].DecorationList {
			wallpaperPath := ""
			iconPath := ""
			switch db.DecorationType {
				case SelectIconName:
					iconPath = decoration.DecorationPath
					wallpaperPath = utils.GetWallpaperPath(currentPath, parentPath)
				case SelectWallpaperName:
					iconPath = utils.GetIconPath(parentPath, currentDirectory.DisplayName)
					wallpaperPath = decoration.DecorationPath
				case SelectListWallpaperName:
					wallpaperPath = decoration.DecorationPath
			}
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     decoration.DecorationName,
				Selected: false,
				Focused:  false,
				Metadata: decoration,
				ImageFilename: iconPath,
				BackgroundFilename: wallpaperPath,
			})
		}
	}
	return menuItems
}

func (db DecorationBrowser) genDirectoryMenuItems() ([]gaba.MenuItem) {
	var menuItems []gaba.MenuItem
	_, decorationAggregation := state.GetDecorationAggregation()
	topLevel := false
	if db.DecorationBrowserIndex == DefaultDecorationBrowserIndex {
		topLevel = true
	}
	if topLevel {
		for index, aggregate := range decorationAggregation {
			menuItems = append(menuItems, gaba.MenuItem{
				Text: aggregate.DirectoryName,
				Selected: false,
				Focused:  false,
				Metadata: index,
			})
		}
	} else {
		_, currentPath, parentPath := utils.GetCurrentDecorationDetails(db.RomDirectoryList)
		currentDirectory := db.RomDirectoryList[len(db.RomDirectoryList) - 1]
		for _, decoration := range decorationAggregation[db.DecorationBrowserIndex].DecorationList {
			wallpaperPath := ""
			iconPath := ""
			switch db.DecorationType {
				case SelectIconName:
					iconPath = decoration.DecorationPath
					wallpaperPath = utils.GetWallpaperPath(currentPath, parentPath)
				case SelectWallpaperName:
					iconPath = utils.GetIconPath(parentPath, currentDirectory.DisplayName)
					wallpaperPath = decoration.DecorationPath
				case SelectListWallpaperName:
					wallpaperPath = decoration.DecorationPath
			}
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     decoration.DecorationName,
				Selected: false,
				Focused:  false,
				Metadata: decoration,
				ImageFilename: iconPath,
				BackgroundFilename: wallpaperPath,
			})
		}
	}
	return menuItems
}
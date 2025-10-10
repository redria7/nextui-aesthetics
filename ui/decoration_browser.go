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
	AggregationOverride		bool
}

func InitDecorationBrowser(romDirectoryList []shared.RomDirectory, listWallpaperSelected bool, decorationType string, decorationbrowserIndex int, aggregationOverride bool) DecorationBrowser {
	overrideLocal := false
	if decorationbrowserIndex != DefaultDecorationBrowserIndex {
		overrideLocal = aggregationOverride
	}
	return DecorationBrowser{
		RomDirectoryList:	romDirectoryList,
		ListWallpaperSelected:	listWallpaperSelected,
		DecorationType: decorationType,
		DecorationBrowserIndex: decorationbrowserIndex,
		AggregationOverride: overrideLocal,
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
	var parentAggName string
	aggregationType := state.GetAppState().Config.DecorationAggregationType
	if aggregationType == utils.AggregateByConsole {
		menuItems, parentAggName = db.genConsoleMenuItems()
	}
	if aggregationType == utils.AggregateByDirectory {
		menuItems, parentAggName = db.genDirectoryMenuItems()
	}

	// Set options
	var decorationTypeName string
	switch db.DecorationType {
		case SelectIconName:
			decorationTypeName = "Icon"
		case SelectWallpaperName:
			decorationTypeName = "Wallpaper"
		case SelectListWallpaperName:
			decorationTypeName = "List Wallpaper"
	}
	title := currentDirectory.DisplayName + ": " + decorationTypeName
	if !topLevel {
		title = parentAggName
	}
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
		selectText = "Open"
		actionText = "Swap Aggregation"
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
			foundIndex := metadata.(int)
			if foundIndex < 0 {
				foundIndex = flipIndex(foundIndex)
				db.AggregationOverride = true
			}
			return foundIndex, exit_code, nil
		}
		return metadata.(models.Decoration), exit_code, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

func (db DecorationBrowser) genConsoleMenuItems() ([]gaba.MenuItem, string) {
	var menuItems []gaba.MenuItem
	parentAggName := ""
	decorationAggregation, _ := state.GetDecorationAggregation()
	topLevel := false
	if db.DecorationBrowserIndex == DefaultDecorationBrowserIndex {
		topLevel = true
	}
	_, currentPath, parentPath := utils.GetCurrentDecorationDetails(db.RomDirectoryList)
	currentDirectory := db.RomDirectoryList[len(db.RomDirectoryList) - 1]
	currentWallpaperPath := utils.GetWallpaperPath(currentPath, parentPath)
	currentIconPath := utils.GetIconPath(parentPath, currentDirectory.DisplayName)
	if db.DecorationType == SelectListWallpaperName {
		currentIconPath = ""
		currentWallpaperPath = utils.GetListWallpaperPath(currentPath)
	}
	if topLevel {
		currentConsole := utils.FindConsoleTag(currentPath)
		var nonCurrentConsoleList []gaba.MenuItem
		for index, aggregate := range decorationAggregation {
			if currentConsole != "" && aggregate.ConsoleTag == currentConsole {
				menuItems = append(menuItems, gaba.MenuItem{
					Text: aggregate.ConsoleName,
					Selected: false,
					Focused:  false,
					Metadata: index,
					ImageFilename: currentIconPath,
					BackgroundFilename: currentWallpaperPath,
				})
			} else {
				nonCurrentConsoleList = append(nonCurrentConsoleList, gaba.MenuItem{
					Text: aggregate.ConsoleName,
					Selected: false,
					Focused:  false,
					Metadata: index,
					ImageFilename: currentIconPath,
					BackgroundFilename: currentWallpaperPath,
				})
			}
		}
		menuItems = append(menuItems, nonCurrentConsoleList...)
	} else {
		parentAggName = decorationAggregation[db.DecorationBrowserIndex].ConsoleName
		for _, decoration := range decorationAggregation[db.DecorationBrowserIndex].DecorationList {
			wallpaperPath := ""
			iconPath := ""
			switch db.DecorationType {
				case SelectIconName:
					iconPath = decoration.DecorationPath
					wallpaperPath = currentWallpaperPath
				case SelectWallpaperName:
					iconPath = currentIconPath
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
	return menuItems, parentAggName
}

func (db DecorationBrowser) genDirectoryMenuItems() ([]gaba.MenuItem, string) {
	var menuItems []gaba.MenuItem
	parentAggName := ""
	consoleAggregation, decorationAggregation := state.GetDecorationAggregation()
	topLevel := false
	if db.DecorationBrowserIndex == DefaultDecorationBrowserIndex {
		topLevel = true
	}
	_, currentPath, parentPath := utils.GetCurrentDecorationDetails(db.RomDirectoryList)
	currentDirectory := db.RomDirectoryList[len(db.RomDirectoryList) - 1]
	currentWallpaperPath := utils.GetWallpaperPath(currentPath, parentPath)
	currentIconPath := utils.GetIconPath(parentPath, currentDirectory.DisplayName)
	if db.DecorationType == SelectListWallpaperName {
		currentIconPath = ""
		currentWallpaperPath = utils.GetListWallpaperPath(currentPath)
	}
	if topLevel {
		currentConsole := utils.FindConsoleTag(currentPath)
		if currentConsole != "" {
			for index, aggregate := range consoleAggregation {
				if aggregate.ConsoleTag == currentConsole {
					menuItems = append(menuItems, gaba.MenuItem{
						Text: aggregate.ConsoleName,
						Selected: false,
						Focused:  false,
						Metadata: flipIndex(index),
						ImageFilename: currentIconPath,
						BackgroundFilename: currentWallpaperPath,
					})
				}
			}
		}
		for index, aggregate := range decorationAggregation {
			menuItems = append(menuItems, gaba.MenuItem{
				Text: aggregate.DirectoryName,
				Selected: false,
				Focused:  false,
				Metadata: index,
				ImageFilename: currentIconPath,
				BackgroundFilename: currentWallpaperPath,
			})
		}
	} else {
		var decorationList []models.Decoration
		if db.AggregationOverride {
			parentAggName = consoleAggregation[db.DecorationBrowserIndex].ConsoleName
			decorationList = consoleAggregation[db.DecorationBrowserIndex].DecorationList
		} else {
			parentAggName = decorationAggregation[db.DecorationBrowserIndex].DirectoryName
			decorationList = decorationAggregation[db.DecorationBrowserIndex].DecorationList
		}
		for _, decoration := range decorationList {
			wallpaperPath := ""
			iconPath := ""
			switch db.DecorationType {
				case SelectIconName:
					iconPath = decoration.DecorationPath
					wallpaperPath = currentWallpaperPath
				case SelectWallpaperName:
					iconPath = currentIconPath
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
	return menuItems, parentAggName
}

func flipIndex(index int) int {
	if index >= 0 {
		return (index * -1) - 1
	}
	return (index + 1) * -1
}

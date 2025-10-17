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
	ClearWallpaperName = 		"Clear Wallpaper"
	SelectWallpaperName = 		"Select Wallpaper"
	ClearListWallpaperName = 	"Clear List Wallpaper"
	SelectListWallpaperName = 	"Select List Wallpaper"
	clearDefaultWallpaperName = 	"Clear Default Wallpaper"
	selectDefaultWallpaperName = 	"Select Default Wallpaper"
	ClearIconName = 			"Clear Icon"
	SelectIconName = 			"Select Icon"
)

type DecorationOptions struct{
	RomDirectoryList		[]shared.RomDirectory
	ListWallpaperSelected	bool
}

func InitDecorationOptions(romDirectoryList []shared.RomDirectory, listWallpaperSelected bool) DecorationOptions {
	return DecorationOptions{
		RomDirectoryList:		romDirectoryList,
		ListWallpaperSelected:	listWallpaperSelected,
	}
}

func (do DecorationOptions) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DecorationOptions
}

func (do DecorationOptions) Draw() (interface{}, int, error) {
	topLevel, currentPath, parentPath := utils.GetCurrentDecorationDetails(do.RomDirectoryList)
	currentDirectory := do.RomDirectoryList[len(do.RomDirectoryList) - 1]
	title := currentDirectory.DisplayName

	// Add items to menu
	var menuItems []gaba.MenuItem
	if do.ListWallpaperSelected {
		title = title + " List Options"
		wallpaperPath := utils.GetListWallpaperPath(currentPath)
		clearName := ClearListWallpaperName
		selectName := SelectListWallpaperName
		if topLevel {
			clearName = clearDefaultWallpaperName
			selectName = selectDefaultWallpaperName
		}
		if utils.CheckListWallpaperPath(currentPath) {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     clearName,
				Selected: false,
				Focused:  false,
				Metadata: ClearListWallpaperName,
				BackgroundFilename: wallpaperPath,
			})
		}
		menuItems = append(menuItems, gaba.MenuItem{
				Text:     selectName,
				Selected: false,
				Focused:  false,
				Metadata: SelectListWallpaperName,
				BackgroundFilename: wallpaperPath,
			})
	} else {
		title = title + " Decoration Options"
		wallpaperPath := utils.GetWallpaperPath(currentPath, parentPath)
		iconPath := utils.GetIconPath(parentPath, utils.GetSimpleFileName(currentDirectory.Path))
		if utils.CheckWallpaperPath(currentPath) {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     ClearWallpaperName,
				Selected: false,
				Focused:  false,
				Metadata: ClearWallpaperName,
				ImageFilename: iconPath,
				BackgroundFilename: wallpaperPath,
			})
		}
		menuItems = append(menuItems, gaba.MenuItem{
				Text:     SelectWallpaperName,
				Selected: false,
				Focused:  false,
				Metadata: SelectWallpaperName,
				ImageFilename: iconPath,
				BackgroundFilename: wallpaperPath,
			})
		if utils.CheckIconPath(parentPath, currentDirectory.DisplayName) {
			menuItems = append(menuItems, gaba.MenuItem{
				Text:     ClearIconName,
				Selected: false,
				Focused:  false,
				Metadata: ClearIconName,
				ImageFilename: iconPath,
				BackgroundFilename: wallpaperPath,
			})
		}
		menuItems = append(menuItems, gaba.MenuItem{
				Text:     SelectIconName,
				Selected: false,
				Focused:  false,
				Metadata: SelectIconName,
				ImageFilename: iconPath,
				BackgroundFilename: wallpaperPath,
			})
	}

	// Set options
	options := gaba.DefaultListOptions(title, menuItems)
	options.SmallTitle = true
	options.EnableImages = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Back"},
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
		return selection.Unwrap().SelectedItem.Metadata.(string), utils.ExitCodeSelect, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

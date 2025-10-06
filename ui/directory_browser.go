package ui

import (
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/filebrowser"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
	"qlova.tech/sum"
)

const (
	CollectionsDisplayName 			= "Collections"
	CollectionsTag         			= "Collections"
	RecentlyPlayedName				= "Recently Played"
	RecentlyPlayedTag				= "Recently Played"
	DefaultListWallpaper			= "List Wallpaper"
	MainListWallpaper				= "Default Wallpaper"
	ExitCodeDefaultListWallpaper	= 5
)

type DirectoryBrowser struct {
	RomDirectoryList		[]shared.RomDirectory
}

func InitDirectoryBrowser(romDirectoryList []shared.RomDirectory) DirectoryBrowser {
	return DirectoryBrowser{
		RomDirectoryList:	romDirectoryList,
	}
}

func (db DirectoryBrowser) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.DirectoryBrowser
}

func (db DirectoryBrowser) Draw() (item interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()
	current_directory := shared.RomDirectory{
		DisplayName: "Main Menu",
		Tag:         "Main Menu",
		Path:        utils.GetRomDirectory(),
	}
	listWallpaperName := MainListWallpaper
	if len(db.RomDirectoryList) > 0 {
		listWallpaperName = DefaultListWallpaper
		current_directory = db.RomDirectoryList[len(db.RomDirectoryList) - 1]
	}

	// Add items to menu
	var menuItems []gaba.MenuItem
	//		Add default list wallpaper option
	menuItems = append(menuItems, gaba.MenuItem{
		Text:     listWallpaperName,
		Selected: false,
		Focused:  false,
		Metadata: DefaultListWallpaper,
		BackgroundFilename: utils.GetListWallpaperPath(current_directory.Path),
	})
	//		If main menu, add collections and recently played
	if len(db.RomDirectoryList) == 0 {
		if collectionsItem := buildCollectionsMenuItem(current_directory, logger); collectionsItem != nil {
			menuItems = append(menuItems, *collectionsItem)
		}
		menuItems = append(menuItems, gaba.MenuItem{
			Text:     RecentlyPlayedName,
			Selected: false,
			Focused:  false,
			Metadata: shared.RomDirectory{
				DisplayName: RecentlyPlayedName,
				Tag:         RecentlyPlayedTag,
				Path:        utils.RecentlyPlayedDirectory,
			},
			ImageFilename: utils.GetIconPath(common.SDCardRoot, RecentlyPlayedName),
			BackgroundFilename: utils.GetWallpaperPath(utils.RecentlyPlayedDirectory, current_directory.Path),
		})
	}
	//		Always add relevant folders
	// romItems, err := buildRomDirectoryMenuItems(current_directory, logger)
	// if err != nil {
	// 	return nil, utils.ExitCodeError, err
	// }
	// menuItems = append(menuItems, romItems...)

	// Set options
	title := current_directory.DisplayName
	options := gaba.DefaultListOptions(title, menuItems)
	options.SmallTitle = true
	options.EnableAction = true
	options.EnableImages = true

	// Set index
	selectedIndex, visibleStartIndex := state.GetCurrentMenuPosition()
	options.SelectedIndex = selectedIndex
	options.VisibleStartIndex = visibleStartIndex

	// Set footers
	options.FooterHelpItems = []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
		{ButtonName: "A", HelpText: "Open Folder"},
		{ButtonName: "X", HelpText: "Decoration Options"},
	}

	// Set Help
	options.EnableHelp = true
	options.HelpTitle = "Directory List Controls"
	options.HelpText = []string{
		"• A: Drill down into the currently selected folder",
		"• X: View decoration options for the current selection",
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
		if metadata == DefaultListWallpaper {
			return nil, ExitCodeDefaultListWallpaper, nil
		}
		if !selection.Unwrap().ActionTriggered {
			exit_code = utils.ExitCodeSelect
		}
		return metadata.(shared.RomDirectory), exit_code, nil
	}

	return nil, utils.ExitCodeCancel, nil
}

func buildCollectionsMenuItem(current_directory shared.RomDirectory, logger *zap.Logger) *gaba.MenuItem {
	fb := filebrowser.NewFileBrowser(logger)

	if err := fb.CWD(utils.GetCollectionDirectory(), false); err != nil {
		logger.Info("Unable to fetch collection directories, skipping", zap.Error(err))
		return nil
	}

	if len(fb.Items) == 0 {
		return nil
	}

	return &gaba.MenuItem{
		Text:     CollectionsDisplayName,
		Selected: false,
		Focused:  false,
		Metadata: shared.RomDirectory{
			DisplayName: CollectionsDisplayName,
			Tag:         CollectionsTag,
			Path:        common.CollectionDirectory,
		},
		ImageFilename: utils.GetIconPath(common.SDCardRoot, CollectionsDisplayName),
		BackgroundFilename: utils.GetWallpaperPath(common.CollectionDirectory, current_directory.Path),
	}
}

func buildRomDirectoryMenuItems(current_directory shared.RomDirectory, logger *zap.Logger) ([]gaba.MenuItem, error) {
	fb := filebrowser.NewFileBrowser(logger)

	// TODO: check user settings for hide empty
	// if err := fb.CWD(utils.GetRomDirectory(), state.GetAppState().Config.HideEmpty); err != nil {
	if err := fb.CWD(current_directory.Path, true); err != nil {
		showRomDirectoryError()
		common.LogStandardFatal("Error fetching directories", err)
		return nil, err
	}

	// Add every non-self-contained-game, non-empty folder to the list
	var menuItems []gaba.MenuItem
	for _, item := range fb.Items {
		if item.IsDirectory && !item.IsSelfContainedDirectory {
			romDirectory := utils.CreateRomDirectoryFromItem(item)
			menuItem := gaba.MenuItem{
				Text:     romDirectory.DisplayName,
				Selected: false,
				Focused:  false,
				Metadata: romDirectory,
				ImageFilename: utils.GetIconPath(current_directory.Path, romDirectory.DisplayName),
				BackgroundFilename: utils.GetWallpaperPath(romDirectory.Path, current_directory.Path),
			}
			menuItems = append(menuItems, menuItem)
		}
	}

	return menuItems, nil
}

func showRomDirectoryError() {
	gaba.ConfirmationMessage("Unable to fetch directories!", []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "Quit"},
	}, gaba.MessageOptions{})
}

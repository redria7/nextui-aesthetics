package main

import (
	_ "github.com/UncleJunVIP/certifiable"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	"go.uber.org/zap"
	"log"
	"os"
	"fmt"
	"time"
	"nextui-aesthetics/state"
	"nextui-aesthetics/models"
	"nextui-aesthetics/utils"
	"nextui-aesthetics/ui"
)

const (
	defaultLogLevel      = "ERROR"
	defaultDirPerm       = 0755
	shortMessageDelay    = 1250 * time.Millisecond
	standardMessageDelay = 2 * time.Second
	longMessageDelay     = 3 * time.Second
)

func init() {
	gaba.InitSDL(gaba.Options{
		WindowTitle:    "Aesthetics",
		ShowBackground: true,
	})

	common.SetLogLevel(defaultLogLevel)
	common.InitIncludes()

	config, err := loadConfig()
	if err != nil {
		log.Fatal("Unable to initialize configuration", zap.Error(err))
	}

	common.SetLogLevel(config.LogLevel)
	state.SetConfig(config)

	logger := common.GetLoggerInstance()
	logger.Debug("Configuration loaded", zap.Object("config", config))

	// collectionDir := utils.GetCollectionDirectory()
	// if _, err := os.Stat(collectionDir); os.IsNotExist(err) {
	// 	if mkdirErr := os.MkdirAll(collectionDir, defaultDirPerm); mkdirErr != nil {
	// 		gaba.ConfirmationMessage("Unable to create Collections directory!", []gaba.FooterHelpItem{
	// 			{ButtonName: "B", HelpText: "Quit"},
	// 		}, gaba.MessageOptions{})
	// 		log.Fatal("Unable to create collection directory", zap.Error(mkdirErr))
	// 	}
	// }
}

func loadConfig() (*models.Config, error) {
	config, err := state.LoadConfig()
	if err != nil {
		config = &models.Config{
			// HideEmpty:       false,
			LogLevel:        defaultLogLevel,
		}
		if saveErr := utils.SaveConfig(config); saveErr != nil {
			return nil, fmt.Errorf("failed to save default config: %w", saveErr)
		}
	}
	return config, nil
}

func main() {
	defer cleanup()

	logger := common.GetLoggerInstance()
	logger.Info("Starting Aesthetics")

	runApplicationLoop()
}

func cleanup() {
	gaba.CloseSDL()
	common.CloseLogger()
}

func runApplicationLoop() {
	var screen models.Screen
	screen = ui.InitMainMenu()
	state.AddNewMenuPosition()

	for {
		result, code, _ := screen.Draw() // TODO: Implement proper error handling
		screen = handleScreenTransition(screen, result, code)
	}
}

func handleScreenTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	switch currentScreen.Name() {
	case models.ScreenNames.MainMenu:
		return handleMainMenuTransition(result, code)
	case models.ScreenNames.DirectoryBrowser:
		return handleDirectoryBrowserTransition(currentScreen, result, code)
	// case models.ScreenNames.Settings:
	// 	state.ReturnToMain()
	// 	return ui.InitMainMenu()
	// case models.ScreenNames.CollectionsList:
	// 	return handleCollectionsListTransition(result, code)
	// case models.ScreenNames.CollectionManagement:
	// 	return handleCollectionManagementTransition(currentScreen, result, code)
	// case models.ScreenNames.CollectionOptions:
	// 	return handleCollectionOptionsTransition(currentScreen, result, code)
	// case models.ScreenNames.Tools:
	// 	return handleToolsTransition(result, code)
	// case models.ScreenNames.GlobalActions:
	// 	return handleGlobalActionsTransition(code)
	// case models.ScreenNames.GamesList:
	// 	return handleGamesListTransition(currentScreen, result, code)
	// case models.ScreenNames.SearchBox:
	// 	return handleSearchBoxTransition(currentScreen, result, code)
	// case models.ScreenNames.Actions:
	// 	return handleActionsTransition(currentScreen, result, code)
	// case models.ScreenNames.BulkActions:
	// 	return handleBulkActionsTransition(currentScreen, result, code)
	// case models.ScreenNames.AddToCollection:
	// 	return handleAddToCollectionTransition(currentScreen, code)
	// case models.ScreenNames.CollectionCreate:
	// 	return handleCollectionCreateTransition(currentScreen)
	// case models.ScreenNames.DownloadArt:
	// 	return handleDownloadArtTransition(currentScreen)
	// case models.ScreenNames.AddToArchive:
	// 	return handleAddToArchiveTransition(currentScreen, result, code)
	// case models.ScreenNames.ArchiveCreate:
	// 	return handleArchiveCreateTransition(currentScreen, result, code)
	// case models.ScreenNames.ArchiveList:
	// 	return handleArchiveListTransition(result, code)
	// case models.ScreenNames.ArchiveGamesList:
	// 	return handleArchiveGamesListTransition(currentScreen, result, code)
	// case models.ScreenNames.ArchiveManagement:
	// 	return handleArchiveManagementTransition(currentScreen, result, code)
	// case models.ScreenNames.ArchiveOptions:
	// 	return handleArchiveOptionsTransition(currentScreen, result, code)
	// case models.ScreenNames.PlayHistoryList:
	// 	return handlePlayHistoryListTransition(currentScreen, result, code)
	// case models.ScreenNames.PlayHistoryGameList:
	// 	return handlePlayHistoryGameListTransition(currentScreen, result, code)
	// case models.ScreenNames.PlayHistoryGameDetails:
	// 	return handlePlayHistoryGameDetailsTransition(currentScreen, result, code)
	// case models.ScreenNames.PlayHistoryGameHistory:
	// 	return handlePlayHistoryGameHistoryTransition(currentScreen, result, code)
	// case models.ScreenNames.PlayHistoryFilter:
	// 	return handlePlayHistoryFilterTransition(currentScreen, result, code)
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

func handleMainMenuTransition(result interface{}, code int) models.Screen {
	switch code {
	case utils.ExitCodeSelect:
		state.AddNewMenuPosition()
		romDir := result.(shared.RomDirectory)
		if romDir.DisplayName == ui.DecorationsDisplayName {
			return ui.InitDirectoryBrowser([]shared.RomDirectory{})
		}
	case utils.ExitCodeError, utils.ExitCodeCancel:
		os.Exit(0)
		return nil
	}
	state.ReturnToMain()
	return ui.InitMainMenu()
}

func handleDirectoryBrowserTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	db := currentScreen.(ui.DirectoryBrowser)

	switch code {
	case utils.ExitCodeSelect:
		state.AddNewMenuPosition()
		return ui.InitDirectoryBrowser(append(db.RomDirectoryList, result.(shared.RomDirectory)))
	case utils.ExitCodeCancel:
		dbListLength := len(db.RomDirectoryList)
		if dbListLength == 0 {
			state.ReturnToMain()
			return ui.InitMainMenu() 
		}
		return ui.InitDirectoryBrowser(db.RomDirectoryList[:dbListLength - 1])
	case utils.ExitCodeAction:
		utils.ShowTimedMessage(fmt.Sprintf("Action on %s!", result.(shared.RomDirectory).DisplayName), shortMessageDelay)
		return ui.InitDirectoryBrowser(db.RomDirectoryList)
	default:
		state.ReturnToMain()
		return ui.InitMainMenu()
	}
}

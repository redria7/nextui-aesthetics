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
		LogFilename:    "aesthetics.log",
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
		case models.ScreenNames.Settings:
			state.ReturnToMain()
			return ui.InitMainMenu()
		case models.ScreenNames.DirectoryBrowser:
			return handleDirectoryBrowserTransition(currentScreen, result, code)
		case models.ScreenNames.DecorationOptions:
			return handleDecorationOptionsTransition(currentScreen, result, code)
		case models.ScreenNames.DecorationBrowser:
			return handleDecorationBrowserTransition(currentScreen, result, code)
		default:
			state.ReturnToMain()
			return ui.InitMainMenu()
	}
}

func handleMainMenuTransition(result interface{}, code int) models.Screen {
	switch code {
		case utils.ExitCodeSelect:
			state.AddNewMenuPosition()
			romDir := result.(string)
			if romDir == ui.DecorationsDisplayName {
				return ui.InitDirectoryBrowser([]shared.RomDirectory{
					shared.RomDirectory{
						DisplayName: "Main Menu",
						Tag:         "Main Menu",
						Path:        utils.GetRomDirectory(),
				}})
			}
		case utils.ExitCodeAction:
			return ui.InitSettingsScreen()
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
			if dbListLength == 1 {
				state.ReturnToMain()
				return ui.InitMainMenu() 
			}
			state.RemoveMenuPositions(1)
			return ui.InitDirectoryBrowser(db.RomDirectoryList[:dbListLength - 1])
		case utils.ExitCodeAction:
			newDirectory := result.(shared.RomDirectory)
			state.AddNewMenuPosition()
			return ui.InitDecorationOptions(append(db.RomDirectoryList, newDirectory), false)
		case ui.ExitCodeDefaultListWallpaper:
			state.AddNewMenuPosition()
			return ui.InitDecorationOptions(db.RomDirectoryList, true)
		default:
			state.ReturnToMain()
			return ui.InitMainMenu()
	}
}

func handleDecorationOptionsTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	do := currentScreen.(ui.DecorationOptions)
	switch code {
		case utils.ExitCodeSelect:
			selectedAction := result.(string)
			switch selectedAction {
				case ui.ClearIconName, ui.ClearWallpaperName, ui.ClearListWallpaperName:
					// Needs activity here -> confirmation screen to delete file
					utils.ShowTimedMessage(fmt.Sprintf("Deleted %s!", selectedAction), shortMessageDelay)
					return ui.InitDecorationOptions(do.RomDirectoryList, do.ListWallpaperSelected)
				case ui.SelectIconName, ui.SelectWallpaperName, ui.SelectListWallpaperName:
					state.AddNewMenuPosition()
					return ui.InitDecorationBrowser(do.RomDirectoryList, do.ListWallpaperSelected, selectedAction, ui.DefaultDecorationBrowserIndex)
				default:
					utils.ShowTimedMessage(fmt.Sprintf("Unsupported action selected %s!\nReport bug please.", selectedAction), shortMessageDelay)
					return ui.InitDecorationOptions(do.RomDirectoryList, do.ListWallpaperSelected)
			}
		default:
			state.RemoveMenuPositions(1)
			if do.ListWallpaperSelected {
				return ui.InitDirectoryBrowser(do.RomDirectoryList)
			}
			return ui.InitDirectoryBrowser(do.RomDirectoryList[:len(do.RomDirectoryList) - 1])
	}
}

func handleDecorationBrowserTransition(currentScreen models.Screen, result interface{}, code int) models.Screen {
	db := currentScreen.(ui.DecorationBrowser)
	if db.DecorationBrowserIndex == ui.DefaultDecorationBrowserIndex {
		switch code {
			case utils.ExitCodeSelect:
				state.AddNewMenuPosition()
				return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, result.(int))
			case utils.ExitCodeAction:
				state.UpdateCurrentMenuPosition(0, 0)
				state.CycleAggregationMode()
				return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, db.DecorationBrowserIndex)
			default:
				state.RemoveMenuPositions(1)
				return ui.InitDecorationOptions(db.RomDirectoryList, db.ListWallpaperSelected)
		}
	}
	switch code {
		case utils.ExitCodeSelect:
			decoration := result.(models.Decoration)
			// Needs activity here -> confirmation screen to copy file
			utils.ShowTimedMessage(fmt.Sprintf("Selected %s!", decoration.DecorationName), shortMessageDelay)
			return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, db.DecorationBrowserIndex)
		case utils.ExitCodeAction:
			decoration := result.(models.Decoration)
			// Needs activity here -> confirmation screen to delete file
			utils.ShowTimedMessage(fmt.Sprintf("Deleted %s!", decoration.DecorationName), shortMessageDelay)
			return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, db.DecorationBrowserIndex)
		default:
			state.RemoveMenuPositions(1)
			return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, ui.DefaultDecorationBrowserIndex)
	}
	
}

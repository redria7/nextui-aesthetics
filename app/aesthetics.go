package main

import (
	"fmt"
	"log"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/ui"
	"nextui-aesthetics/utils"
	"os"
	"strings"
	"time"

	_ "github.com/UncleJunVIP/certifiable"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	shared "github.com/UncleJunVIP/nextui-pak-shared-functions/models"
	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"go.uber.org/zap"
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
					currentDirectory := do.RomDirectoryList[len(do.RomDirectoryList) - 1]
					_, currentPath, parentPath := utils.GetCurrentDecorationDetails(do.RomDirectoryList)
					destinationPath := ""
					switch selectedAction {
						case ui.ClearIconName:
							destinationPath = utils.GetTrueIconPath(parentPath, currentDirectory.DisplayName)
						case ui.ClearWallpaperName:
							destinationPath = utils.GetTrueWallpaperPath(currentPath)
						case ui.ClearListWallpaperName:
							destinationPath = utils.GetTrueListWallpaperPath(currentPath)
					}
					if confirmDeletion("Clear this decoration from:\n" + splitPathToLines(destinationPath), destinationPath) {
						state.UpdateCurrentMenuPosition(0, 0)
						res := common.DeleteFile(destinationPath)
						if res {
							utils.ShowTimedMessage(fmt.Sprintf("Deleted:\n%s", splitPathToLines(destinationPath)), shortMessageDelay)
						} else {
							utils.ShowTimedMessage(fmt.Sprintf("Failed to delete:%s", splitPathToLines(destinationPath)), shortMessageDelay)
						}
					}
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
			return copyFile(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, db.DecorationBrowserIndex, result.(models.Decoration))
		case utils.ExitCodeAction:
			decoration := result.(models.Decoration)
			if confirmDeletion("Delete this decoration from:\n" + splitPathToLines(decoration.DecorationPath), decoration.DecorationPath) {
				res := common.DeleteFile(decoration.DecorationPath)
				if res {
					// Successful file deletion. Clear from aggregations
					// TODO: Possibly run additional logic checks to change selected item state?
					consoleAggregation, decorationAggregation := state.GetDecorationAggregation()
					shouldBreak := false
					for index, agg := range consoleAggregation {
						if decoration.ConsoleName == agg.ConsoleName {
							for subIndex, dec := range agg.DecorationList {
								if decoration.DecorationName == dec.DecorationName {
									agg.DecorationList = append(agg.DecorationList[:subIndex], agg.DecorationList[subIndex+1:]...)
									consoleAggregation[index] = agg
									if len(agg.DecorationList) == 0 {
										consoleAggregation = append(consoleAggregation[:index], consoleAggregation[index+1:]...)
									}
									shouldBreak = true
									break
								}
							}
							if shouldBreak {
								break
							}
						}
					}
					shouldBreak = false
					for index, agg := range decorationAggregation {
						if decoration.DirectoryName == agg.DirectoryName {
							for subIndex, dec := range agg.DecorationList {
								if decoration.DecorationName == dec.DecorationName {
									agg.DecorationList = append(agg.DecorationList[:subIndex], agg.DecorationList[subIndex+1:]...)
									decorationAggregation[index] = agg
									if len(agg.DecorationList) == 0 {
										decorationAggregation = append(decorationAggregation[:index], decorationAggregation[index+1:]...)
									}
									shouldBreak = true
									break
								}
							}
							if shouldBreak {
								break
							}
						}
					}
					utils.ShowTimedMessage(fmt.Sprintf("Deleted:\n%s", splitPathToLines(decoration.DecorationPath)), shortMessageDelay)
				} else {
					utils.ShowTimedMessage(fmt.Sprintf("Failed to delete:%s", splitPathToLines(decoration.DecorationPath)), shortMessageDelay)
				}
			}
			return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, db.DecorationBrowserIndex)
		default:
			state.RemoveMenuPositions(1)
			return ui.InitDecorationBrowser(db.RomDirectoryList, db.ListWallpaperSelected, db.DecorationType, ui.DefaultDecorationBrowserIndex)
	}
}

func copyFile(romDirectoryList []shared.RomDirectory, listWallpaperSelected bool, decorationType string, decorationBrowserIndex int, decoration models.Decoration) models.Screen {
	currentDirectory := romDirectoryList[len(romDirectoryList) - 1]
	_, currentPath, parentPath := utils.GetCurrentDecorationDetails(romDirectoryList)
	sourcePath := decoration.DecorationPath
	destinationPath := ""
	utils.CheckIconPath(parentPath, currentDirectory.DisplayName)
	utils.CheckWallpaperPath(currentPath)
	utils.CheckListWallpaperPath(currentPath)
	switch decorationType {
		case ui.SelectIconName:
			destinationPath = utils.GetTrueIconPath(parentPath, currentDirectory.DisplayName)
		case ui.SelectWallpaperName:
			destinationPath = utils.GetTrueWallpaperPath(currentPath)
		case ui.SelectListWallpaperName:
			destinationPath = utils.GetTrueListWallpaperPath(currentPath)
	}
	// message := "Copy image from:\n" + splitPathToLines(sourcePath) + "\nto\n" + splitPathToLines(destinationPath)
	message := "Copy image to:\n" + splitPathToLines(destinationPath)
	if utils.ConfirmAction(message, sourcePath) {
		err := utils.CopyFile(sourcePath, destinationPath)
		if err != nil {
			utils.ShowTimedMessage("Unable to copy image!", longMessageDelay)
			return ui.InitDecorationBrowser(romDirectoryList, listWallpaperSelected, decorationType, decorationBrowserIndex)
		}
		utils.ShowTimedMessage("Image copied successfully!", shortMessageDelay)
		state.RemoveMenuPositions(2)
		return ui.InitDecorationOptions(romDirectoryList, listWallpaperSelected)
	}
	return ui.InitDecorationBrowser(romDirectoryList, listWallpaperSelected, decorationType, decorationBrowserIndex)
}

func splitPathToLines(filePath string) string {
	splitList := strings.Split(filePath, "/")
	widthList := []string{""}
	for _, splitPhrase := range(splitList) {
		if splitPhrase != "" {
			currentPhrase := widthList[len(widthList) - 1]
			newPhrase := ""
			if currentPhrase == "" {
				currentPhrase = splitPhrase
			} else {
				curLen := len(currentPhrase)
				newLen := len(splitPhrase)
				if curLen + newLen > 40 {
					currentPhrase = currentPhrase + "/"
					newPhrase = splitPhrase
				} else {
					currentPhrase = currentPhrase + "/" + splitPhrase
				}
			}
			widthList[len(widthList) - 1] = currentPhrase
			if newPhrase != "" {
				widthList = append(widthList, newPhrase)
			}
		}
	}
	return strings.Join(widthList, "\n")
}

func confirmDeletion(message, imagePath string) bool {
	result, err := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: "I Changed My Mind"},
		{ButtonName: "A", HelpText: "Trash It!"},
	}, gaba.MessageOptions{
		ImagePath: imagePath,
	})

	return err == nil && result.IsSome()
}

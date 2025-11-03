package state

import (
	"go.uber.org/atomic"
	"gopkg.in/yaml.v3"
	"fmt"
	"os"
	"sync"
	"nextui-aesthetics/models"
	"nextui-aesthetics/utils"
)

var appState atomic.Pointer[models.AppState]
var onceAppState sync.Once

func LoadConfig() (*models.Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return nil, fmt.Errorf("reading config.yml: %w", err)
	}

	var config models.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("parsing config.yml: %w", err)
	}

	return &config, nil
}

func GetAppState() *models.AppState {
	onceAppState.Do(func() {
		appState.Store(&models.AppState{})
	})
	return appState.Load()
}

func CycleAggregationMode() {
	temp := GetAppState()
	switch temp.Config.DecorationAggregationType {
		case utils.AggregateByConsole:
			temp.Config.DecorationAggregationType = utils.AggregateByDirectory
		case utils.AggregateByDirectory:
			temp.Config.DecorationAggregationType = utils.AggregateByConsole
	}
	UpdateAppState(temp)
}

func UpdateAppState(newAppState *models.AppState) {
	appState.Store(newAppState)
}

func SetConfig(config *models.Config) {
	temp := GetAppState()
	temp.Config = config

	UpdateAppState(temp)
}

func AddNewMenuPosition() {
	temp := GetAppState()
	temp.MenuPositionList = append(temp.MenuPositionList, models.MenuPositionPointer{
		SelectedIndex:    0,
		SelectedPosition: 0,
	})
	UpdateAppState(temp)
}

func UpdateCurrentMenuPosition(newIndex int, newPosition int) {
	temp := GetAppState()
	temp.MenuPositionList[len(temp.MenuPositionList)-1] = models.MenuPositionPointer{
		SelectedIndex:    newIndex,
		SelectedPosition: newPosition,
	}
	UpdateAppState(temp)
}

func RemoveMenuPositions(positionCount int) {
	temp := GetAppState()
	listLength := len(temp.MenuPositionList)

	EndPosition := 0
	if positionCount < 0 {
		EndPosition = -1 * positionCount
		if EndPosition > listLength {
			return
		}
	} else {
		if positionCount > listLength {
			positionCount = listLength
		}
		EndPosition = listLength - positionCount
	}

	temp.MenuPositionList = temp.MenuPositionList[:EndPosition]
	UpdateAppState(temp)
}

func ReturnToMain() {
	RemoveMenuPositions(-1)
}

// func ReturnToArchiveManagement() {
// 	RemoveMenuPositions(-3)
// }

// func ReturnToCollectionManagement() {
// 	RemoveMenuPositions(-3)
// }

func GetCurrentMenuPosition() (int, int) {
	tempList := GetAppState().MenuPositionList
	if len(tempList) <= 0 {
		AddNewMenuPosition()
		tempList = GetAppState().MenuPositionList
	}

	currentPosition := tempList[len(tempList)-1]
	selectedIndex := currentPosition.SelectedIndex
	selectedPosition := currentPosition.SelectedPosition

	selectedPosition = max(0, selectedIndex-selectedPosition)

	return selectedIndex, selectedPosition
}

func GetDecorationAggregation() ([]models.ConsoleAggregation, []models.DirectoryAggregation) {
	temp := GetAppState()
	if temp.DecorationsAggregatedOnConsoles == nil {
		updateDecorationAggregations()
		temp = GetAppState()
	}
	return temp.DecorationsAggregatedOnConsoles, temp.DecorationsAggregatedOnDirectories
}

func updateDecorationAggregations() {
	temp := GetAppState()
	temp.DecorationsAggregatedOnConsoles, temp.DecorationsAggregatedOnDirectories = utils.GenerateDecorationAggregations()
	UpdateAppState(temp)
}

func ClearDecorationAggregations() {
	temp := GetAppState()
	temp.DecorationsAggregatedOnConsoles = nil
	temp.DecorationsAggregatedOnDirectories = nil
	UpdateAppState(temp)
}

func GetThemeCatalog() []models.ThemeSummary {
	temp := GetAppState()
	if temp.ThemeCatalog == nil {
		UpdateThemeCatalog()
		temp = GetAppState()
	}
	return temp.ThemeCatalog
}

func UpdateThemeCatalog() {
	temp := GetAppState()
	themeCatalog := utils.GenerateThemeCatalog()
	currentThemes := utils.GetDownloadedThemes()
	// If an item on the list is not currently downloaded, or if it has no preview.png, then pull the preview.png file
	var themesNeedingPreviews []models.ThemeSummary
	for index, theme := range themeCatalog {
		themeStatus, exists := currentThemes[theme.ThemeName]
		if exists {
			if !themeStatus.PreviewFound {
				themesNeedingPreviews = append(themesNeedingPreviews, theme)
				themeCatalog[index].IsNew = true
			}
		} else {
			themesNeedingPreviews = append(themesNeedingPreviews, theme)
			themeCatalog[index].IsNew = true
		}
	}
	utils.DownloadThemePreviews(themesNeedingPreviews)
	temp.ThemeCatalog = themeCatalog
	UpdateAppState(temp)
}

// func GetPlayMaps() (map[string][]models.PlayHistoryAggregate, map[string]int, int) {
// 	temp := GetAppState()
// 	if temp.GamePlayMap == nil {
// 		updatePlayMaps()
// 		temp = GetAppState()
// 	}
// 	return temp.GamePlayMap, temp.ConsolePlayMap, temp.TotalPlay
// }

// func updatePlayMaps() {
// 	temp := GetAppState()
// 	temp.GamePlayMap, temp.ConsolePlayMap, temp.TotalPlay = utils.GenerateCurrentGameStats("")
// 	UpdateAppState(temp)
// }

// func ClearPlayMaps() {
// 	temp := GetAppState()
// 	temp.GamePlayMap = nil
// 	temp.ConsolePlayMap = nil
// 	temp.TotalPlay = 0
// 	UpdateAppState(temp)
// }

// func GetCollectionMap() map[string][]models.Collection {
// 	temp := GetAppState()
// 	if temp.CollectionMap == nil {
// 		updateCollectionMap()
// 		temp = GetAppState()
// 	}
// 	return temp.CollectionMap
// }

// func updateCollectionMap() {
// 	temp := GetAppState()
// 	temp.CollectionMap = utils.GenerateCollectionMap()
// 	UpdateAppState(temp)
// }

// func ClearCollectionMap() {
// 	temp := GetAppState()
// 	temp.CollectionMap = nil
// 	UpdateAppState(temp)
// }

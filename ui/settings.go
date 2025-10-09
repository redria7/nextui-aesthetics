package ui

import (
	"fmt"
	"github.com/redria7/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"nextui-aesthetics/models"
	"nextui-aesthetics/state"
	"nextui-aesthetics/utils"
	"qlova.tech/sum"
)

type SettingsScreen struct {
}

func InitSettingsScreen() SettingsScreen {
	return SettingsScreen{}
}

func (s SettingsScreen) Name() sum.Int[models.ScreenName] {
	return models.ScreenNames.Settings
}

func (s SettingsScreen) Draw() (settings interface{}, exitCode int, e error) {
	logger := common.GetLoggerInstance()

	appState := state.GetAppState()

	items := []gabagool.ItemWithOptions{
		{
			Item: gabagool.MenuItem{
				Text: "Log Level",
			},
			Options: []gabagool.Option{
				{DisplayName: "Debug", Value: "DEBUG"},
				{DisplayName: "Error", Value: "ERROR"},
			},
			SelectedOption: func() int {
				switch appState.Config.LogLevel {
				case "DEBUG":
					return 0
				case "ERROR":
					return 1
				}
				return 0
			}(),
		},
		{
			Item: gabagool.MenuItem{
				Text: "Decoration Aggregation",
			},
			Options: []gabagool.Option{
				{DisplayName: "On Console", Value: utils.AggregateByConsole},
				{DisplayName: "On Directory", Value: utils.AggregateByDirectory},
			},
			SelectedOption: func() int {
				switch appState.Config.LogLevel {
				case "CONSOLE":
					return 0
				case "DIRECTORY":
					return 1
				}
				return 0
			}(),
		},
	}

	footerHelpItems := []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: "Cancel"},
		{ButtonName: "←→", HelpText: "Cycle"},
		{ButtonName: "Start", HelpText: "Save"},
	}

	result, err := gabagool.OptionsList(
		"Aesthetics Settings",
		items,
		footerHelpItems,
	)

	if err != nil {
		fmt.Println("Error showing options list:", err)
		return
	}

	if result.IsSome() {
		newSettingOptions := result.Unwrap().Items

		for _, option := range newSettingOptions {
			if option.Item.Text == "Log Level" {
				logLevelValue := option.Options[option.SelectedOption].Value.(string)
				appState.Config.LogLevel = logLevelValue
			} else if option.Item.Text == "Decoration Aggregation" {
				decorationAggregationValue := option.Options[option.SelectedOption].Value.(int)
				appState.Config.DecorationAggregationType = decorationAggregationValue
			}
		}

		err := utils.SaveConfig(appState.Config)
		if err != nil {
			logger.Error("Error saving config", zap.Error(err))
			return nil, 0, err
		}

		state.UpdateAppState(appState)

		return result, 0, nil
	}

	return nil, 2, nil
}

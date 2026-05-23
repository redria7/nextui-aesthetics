package ui

import (
	"fmt"
	"github.com/redria7/gabagool/pkg/gabagool"
	"github.com/UncleJunVIP/nextui-pak-shared-functions/common"
	"go.uber.org/zap"
	"nextui-aesthetics/models"
	"nextui-aesthetics/internal/i18n"
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
				Text: i18n.T("ae.settings.log_level"),
			},
			Options: []gabagool.Option{
				{DisplayName: i18n.T("ae.settings.log_level.debug"), Value: "DEBUG"},
				{DisplayName: i18n.T("ae.settings.log_level.error"), Value: "ERROR"},
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
				Text: i18n.T("ae.settings.decoration_aggregation"),
			},
			Options: []gabagool.Option{
				{DisplayName: i18n.T("ae.settings.aggregation.directory"), Value: utils.AggregateByDirectory},
				{DisplayName: i18n.T("ae.settings.aggregation.console"), Value: utils.AggregateByConsole},
			},
			SelectedOption: func() int {
				switch appState.Config.LogLevel {
				case "DIRECTORY":
					return 0
				case "CONSOLE":
					return 1
				}
				return 0
			}(),
		},
	}

	footerHelpItems := []gabagool.FooterHelpItem{
		{ButtonName: "B", HelpText: i18n.T("ae.btn.cancel")},
		{ButtonName: "←→", HelpText: i18n.T("ae.btn.cycle")},
		{ButtonName: "Start", HelpText: i18n.T("ae.btn.save")},
	}

	result, err := gabagool.OptionsList(
		i18n.T("ae.title.settings"),
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
			if option.Item.Text == i18n.T("ae.settings.log_level") {
				logLevelValue := option.Options[option.SelectedOption].Value.(string)
				appState.Config.LogLevel = logLevelValue
			} else if option.Item.Text == i18n.T("ae.settings.decoration_aggregation") {
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

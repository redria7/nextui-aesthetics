package utils

import (
	"time"

	gaba "github.com/redria7/gabagool/pkg/gabagool"
	"nextui-aesthetics/internal/i18n"
)

func ShowTimedMessage(message string, delay time.Duration) {
	gaba.ProcessMessage(message, gaba.ProcessMessageOptions{}, func() (interface{}, error) {
		time.Sleep(delay)
		return nil, nil
	})
}

func ConfirmAction(message string, imagePath string) bool {
	result, err := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: i18n.T("ae.btn.changed_mind")},
		{ButtonName: "A", HelpText: i18n.T("ae.btn.yes")},
	}, gaba.MessageOptions{
		ImagePath: imagePath,
	})

	return err == nil && result.IsSome()
}

func ConfirmActionCustomBack(message string, imagePath string, backText string) bool {
	result, err := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: backText},
		{ButtonName: "A", HelpText: i18n.T("ae.btn.yes")},
	}, gaba.MessageOptions{
		ImagePath: imagePath,
	})

	return err == nil && result.IsSome()
}

func ConfirmBulkAction(message string) bool {
	confirm, _ := gaba.ConfirmationMessage(message, []gaba.FooterHelpItem{
		{ButtonName: "B", HelpText: i18n.T("ae.btn.cancel")},
		{ButtonName: "X", HelpText: i18n.T("ae.btn.remove")},
	}, gaba.MessageOptions{
		ImagePath:     "",
		ConfirmButton: gaba.ButtonX,
	})

	return confirm.IsSome() && !confirm.Unwrap().Cancelled
}

package prompt

import (
	"github.com/manifoldco/promptui"
	"os"
)

type selectPromptUi struct{}

func (receiver selectPromptUi) Yml(label string, items []string) string {
	prompt := promptui.Select{
		Label:    label,
		Items:    items,
		HideHelp: true,
		Size:     10,
	}
	_, result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		os.Exit(1)
	}
	return result
}

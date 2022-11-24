package prompt

import (
	"github.com/manifoldco/promptui"
	"os"
)

type inputPromptUi struct{}

func (receiver inputPromptUi) Run(exitFunc func(), prompt promptui.Prompt, retry ...bool) string {
	result, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" {
			exitFunc()
			os.Exit(0)
		}
		if len(retry) > 0 && retry[0] {
			return receiver.Run(exitFunc, prompt, retry...)
		}
	}
	return result
}

func (receiver inputPromptUi) RunWithLabel(exitFunc func(), label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}
	return receiver.Run(exitFunc, prompt)
}

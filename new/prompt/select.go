package prompt

import (
	"github.com/manifoldco/promptui"
	"os"
)

type selectPromptUi struct{}

func (receiver selectPromptUi) Run(exitFunc func(), prompt promptui.Select) string {
	_, result, err := prompt.Run()
	if err != nil {
		if err.Error() == "^C" {
			exitFunc()
			os.Exit(0)
		}
	}
	return result
}

package prompt

import (
	"github.com/manifoldco/promptui"
)

type inputPromptUi struct{}

func (receiver inputPromptUi) Yml(label string, validate ...promptui.ValidateFunc) string {
	if len(validate) == 0 {
		validate = append(validate, func(input string) error {
			return nil
		})
	}
	prompt := promptui.Prompt{
		Label:       label,
		Validate:    validate[0],
		HideEntered: true,
	}
	result, _ := prompt.Run()
	return result
}

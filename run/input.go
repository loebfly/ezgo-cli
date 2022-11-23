package run

import "github.com/manifoldco/promptui"

type inputPromptUi struct{}

func (receiver inputPromptUi) searchKeyword() string {
	prompt := promptui.Prompt{
		Label: "请输入项目关键字(可为空)",
		Validate: func(input string) error {
			return nil
		},
	}
	result, _ := prompt.Run()
	return result
}

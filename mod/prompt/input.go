package prompt

import (
	"fmt"
	"github.com/manifoldco/promptui"
)

type inputPromptUi struct{}

func (receiver inputPromptUi) SearchKeyword() string {
	prompt := promptui.Prompt{
		Label: "请输入项目关键字(可为空)",
		Validate: func(input string) error {
			return nil
		},
		HideEntered: true,
	}
	result, _ := prompt.Run()
	fmt.Println("您输入的关键字是: " + result)
	return result
}

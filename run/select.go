package run

import "github.com/manifoldco/promptui"

type selectPromptUi struct{}

func (receiver selectPromptUi) Project(items []string) string {
	prompt := promptui.Select{
		Label: "请选择项目",
		Items: items,
	}
	_, result, _ := prompt.Run()
	return result
}

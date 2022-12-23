package prompt

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
)

type selectPromptUi struct{}

func (receiver selectPromptUi) Project(items []string) string {
	prompt := promptui.Select{
		Label:    "请选择项目",
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

func (receiver selectPromptUi) Yml(items []string) string {
	prompt := promptui.Select{
		Label:    "请选择配置文件",
		Items:    items,
		HideHelp: true,
		Size:     10,
	}
	_, result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		os.Exit(1)
	}
	fmt.Println(prompt.Label)
	fmt.Println("您选择的是:", result)
	return result
}

func (receiver selectPromptUi) IsAgree(label string) bool {
	prompt := promptui.Select{
		Label:        label,
		Items:        []string{"是", "否"},
		HideHelp:     true,
		HideSelected: true,
	}
	_, result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		os.Exit(1)
	}
	fmt.Println(label)
	fmt.Println("您选择的是:", result)
	return result == "是"
}

func (receiver selectPromptUi) GoVersion() string {
	prompt := promptui.Select{
		Label:    "请选择Go版本",
		Items:    []string{"1.17", "1.19"},
		HideHelp: true,
		Size:     10,
	}
	_, result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		os.Exit(1)
	}
	fmt.Println(prompt.Label)
	fmt.Println("您选择的是:", result)
	return result
}

func (receiver selectPromptUi) ProjectGroup() string {
	prompt := promptui.Select{
		Label:    "请选择项目组",
		Items:    []string{"无分组", "opencloud", "cmp"},
		HideHelp: true,
		Size:     10,
	}
	_, result, err := prompt.Run()
	if err != nil && err.Error() == "^C" {
		os.Exit(1)
	}
	fmt.Println(prompt.Label)
	fmt.Println("您选择的是:", result)
	return result
}

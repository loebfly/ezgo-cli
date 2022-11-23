package run

import (
	"bytes"
	"ezgo-cli/cmd"
	"ezgo-cli/config"
	"ezgo-cli/let"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var (
	inputUi  = inputPromptUi{}
	selectUi = selectPromptUi{}
)

func Exec() {
	fmt.Println("项目根目录: ", config.GetProjectDir())
	fmt.Println("日志根目录: ", config.GetLogDir())

	projectName := getProjectName()

	projectDir := fmt.Sprintf("%s%s", config.GetProjectDir(), projectName)
	if selectUi.IsAgree("是否要生成swag文档?") {
		fmt.Println("开始执行 swag init")
		_, err := cmd.ExecInDir(projectDir, "swag", "init")
		if err != nil {
			fmt.Printf("生成swag文档失败: %s", err.Error())
			os.Exit(0)
		}
		fmt.Println("swag init 执行完毕")
	}

	if selectUi.IsAgree("是否要编译项目?") {
		_, err := os.Stat(projectDir + "/go.mod")
		if err == nil || os.IsExist(err) {
			fmt.Println("开始执行 go mod tidy")
			_, err := cmd.ExecInDir(projectDir, "go", "mod", "tidy", "-compat=1.17")
			if err != nil {
				fmt.Printf("go mod tidy失败: %s", err.Error())
				os.Exit(0)
			}
			fmt.Println("go mod tidy 执行完毕")
		}
		fmt.Println("开始执行 go build")
		_, err = cmd.ExecInDir(projectDir, "go", "build")
		if err != nil {
			fmt.Printf("编译项目失败: %s", err.Error())
			os.Exit(0)
		}
		fmt.Println("go build 执行完毕")
	}

	if selectUi.IsAgree("是否要后台运行项目?") {
		ymlName := getYmlName(projectDir)
		if ymlName == "" {
			if !selectUi.IsAgree("未找到项目YML配置, 是否继续运行?") {
				os.Exit(0)
			}
		}
		fmt.Printf("开始查找是否有%s配置的旧进程\n", ymlName)
		pidCmd := getOldPidCmd(projectName, ymlName)
		if pidCmd != nil {
			wcCmd, _ := cmd.ExecWithPreCmd(pidCmd, "wc", "-l")
			pidCount := wcCmd.Stdout.(*bytes.Buffer).String()
			pidCount = strings.TrimSpace(pidCount)
			if pidCount != "0" && pidCount != "" {
				fmt.Printf("找到%s配置的%s个旧进程\n", ymlName, pidCount)
				fmt.Printf("开始杀死%s配置的旧进程\n", ymlName)
				pidCmd = getOldPidCmd(projectName, ymlName)
				printCmd, err := cmd.ExecWithPreCmd(pidCmd, "awk", "{print $2}")
				if err != nil {
					fmt.Printf("杀死旧进程失败: %s\n", err.Error())
					os.Exit(0)
				}
				_, err = cmd.ExecWithPreCmd(printCmd, "xargs", "kill", "-9")
				if err != nil {
					fmt.Printf("杀死旧进程失败: %s\n", err.Error())
					os.Exit(0)
				}
				fmt.Printf("杀死%s配置的旧进程成功\n", ymlName)
			} else {
				fmt.Printf("未找到%s配置的旧进程, 无需终止\n", ymlName)
			}
		} else {
			fmt.Printf("未找到%s配置的旧进程, 无需终止\n", ymlName)
		}

		fmt.Println("开启程序后台运行")
		appPath := fmt.Sprintf("%s/%s", projectDir, projectName)
		fmt.Printf("程序路径: %s\n", appPath)
		fmt.Printf("配置路径: %s\n", projectDir+"/"+ymlName)
		_, _ = cmd.ExecWithNohup(appPath, "-f", ymlName, ">", config.GetLogDir(), "2>&1", "&")
		fmt.Printf("查看日志: tail -f -n200 %s%s.$(date %%F).log\n", config.GetLogDir(), projectName)
		fmt.Printf("查看out: tail -f -n200 %s%s.out\n", config.GetLogDir(), projectName)
		fmt.Printf("查看进程: ps -ef | grep %s\n", projectName)
		os.Exit(0)
	}
}

func getProjectName() string {
	keyword := inputUi.searchKeyword()

	if keyword != "" {
		fmt.Printf("正在匹配包含'%s'的项目\n", keyword)
	}

	dirFiles, err := ioutil.ReadDir(config.YmlData.String(let.ProjectDirKey))
	if err != nil {
		fmt.Printf("读取项目目录失败: %s", err.Error())
		os.Exit(0)
	}

	var projects []string
	for _, dirFile := range dirFiles {
		if keyword != "" && !strings.Contains(dirFile.Name(), keyword) {
			continue
		}
		if dirFile.IsDir() {
			projects = append(projects, dirFile.Name())
		}
	}

	if len(projects) == 0 {
		if selectUi.IsAgree("未找到匹配的项目, 是否重新搜索?") {
			return getProjectName()
		}
		os.Exit(0)
	}

	if len(projects) == 1 {
		fmt.Printf("找到唯一匹配的项目: %s\n", projects[0])
		return projects[0]
	}

	return selectUi.Project(projects)
}

func getYmlName(projectDir string) string {
	dirFiles, err := ioutil.ReadDir(projectDir)
	if err != nil {
		fmt.Printf("读取项目YML配置失败: %s", err.Error())
		os.Exit(0)
	}
	var ymlNames []string
	for _, dirFile := range dirFiles {
		if strings.HasSuffix(dirFile.Name(), ".yml") {
			ymlNames = append(ymlNames, dirFile.Name())
		}
	}
	if len(ymlNames) == 0 {
		fmt.Printf("未找到项目YML配置")
		os.Exit(0)
	}
	if len(ymlNames) == 1 {
		return ymlNames[0]
	}
	return selectUi.Yml(ymlNames)
}

func getOldPidCmd(projectName, ymlName string) *exec.Cmd {
	psCmd, err := cmd.ExecInDir("", "ps", "-ef")
	if err != nil {
		return nil
	}

	grepCmd, err := cmd.ExecWithPreCmd(psCmd, "grep", config.GetProjectDir())
	if err != nil {
		return nil
	}

	grepCmd, err = cmd.ExecWithPreCmd(grepCmd, "grep", projectName+"/"+projectName)
	if err != nil {
		return nil
	}

	grepCmd, err = cmd.ExecWithPreCmd(grepCmd, "grep", ymlName)
	if err != nil {
		return nil
	}
	return grepCmd
}

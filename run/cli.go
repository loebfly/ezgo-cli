package run

import (
	"bytes"
	"ezgo-cli/cmd"
	"ezgo-cli/run/prompt"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var (
	OptionsWorkDir   = ""     // -workDir 项目根目录
	OptionsLogDir    = ""     // -logDir 日志目录
	OptionsSwagInit  = "true" // -swagInit 是否初始化swagger
	OptionsGoBuild   = "true" // -build 是否编译
	OptionsGoVersion = "1.17" // -goVersion 指定go版本
	OptionsGroup     = ""     // -group 指定项目组
)

func Exec() {
	cmdFlag := flag.NewFlagSet(cmd.Run, flag.ExitOnError)
	cmdFlag.StringVar(&OptionsWorkDir, "workDir", "/opt/go/src/flamecloud.cn/", "项目根目录")
	cmdFlag.StringVar(&OptionsLogDir, "logDir", "/opt/logs/", "日志根目录")
	cmdFlag.StringVar(&OptionsSwagInit, "swag", "true", "是否生成swagger文档")
	cmdFlag.StringVar(&OptionsGoBuild, "build", "true", "是否要编译项目")
	err := cmdFlag.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("解析命令行参数失败: ", err.Error())
		return
	}

	OptionsGroup = prompt.SelectUi.ProjectGroup()
	if OptionsGroup != "无分组" {
		OptionsWorkDir = OptionsWorkDir + OptionsGroup + "/"
	}

	fmt.Println("项目根目录: ", OptionsWorkDir)
	fmt.Println("日志根目录: ", OptionsLogDir)

	// 选择Go版本
	OptionsGoVersion = prompt.SelectUi.GoVersion()

	// 设置GO_VERSION=1.19
	fmt.Printf("设置GO_VERSION=%s\n", OptionsGoVersion)
	err = os.Setenv("GO_VERSION", OptionsGoVersion)
	if err != nil {
		fmt.Printf("设置GO_VERSION失败: %s", err.Error())
		os.Exit(0)
		return
	}
	projectName := getProjectName()

	projectDir := fmt.Sprintf("%s%s", OptionsWorkDir, projectName)

	if OptionsGoBuild == "true" {
		_, err = os.Stat(projectDir + "/go.mod")
		if err == nil || os.IsExist(err) {
			fmt.Println("开始执行 go mod tidy")
			_, err := cmd.ExecInDirWithPrint(projectDir, "go", "mod", "tidy", "-compat="+OptionsGoVersion)
			if err != nil {
				fmt.Printf("go mod tidy 执行失败: %s\n", err.Error())
				os.Exit(0)
			}
			fmt.Println("go mod tidy 执行完毕")
		}
	}

	if OptionsSwagInit == "true" {
		fmt.Println("开始执行 swag init")
		if OptionsGoVersion == "1.17" {
			_, err := cmd.ExecInDirWithPrint(projectDir, "swag", "init")
			if err != nil {
				fmt.Printf("生成swag文档失败: %s", err.Error())
			}
		} else {
			//_, err := cmd.ExecInDirWithPrint(projectDir, "swag", "init", "--parseDependency", "--parseInternal")
			//if err != nil {
			//	fmt.Printf("生成swag文档失败: %s", err.Error())
			//}
		}

		fmt.Println("swag init 执行完毕")
	}

	if OptionsGoBuild == "true" {
		fmt.Println("开始执行 go build")
		_, err = cmd.ExecInDirWithPrint(projectDir, "go", "build")
		if err != nil {
			fmt.Printf("编译项目失败: %s\n", err.Error())
			os.Exit(0)
		}

		fmt.Println("go build 执行完毕")
	}

	ymlName := getYmlName(projectDir)
	if ymlName == "" {
		if !prompt.SelectUi.IsAgree("未找到项目YML配置, 是否继续运行?") {
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
			_, err = cmd.ExecWithPreCmd(printCmd, "xargs", "kill")
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
	outPath := fmt.Sprintf("%s%s.out", OptionsLogDir, projectName)
	fmt.Printf("程序路径: %s\n", appPath)
	fmt.Printf("配置路径: %s\n", projectDir+"/"+ymlName)

	nohup := fmt.Sprintf("nohup %s -f %s >%s 2>&1 &", appPath, ymlName, outPath)
	_, err = exec.Command("sh", "-c", nohup).CombinedOutput()
	if err != nil {
		fmt.Printf("后台运行项目失败: %s\n", err.Error())
	} else {
		fmt.Println("后台运行项目成功")
	}
	fmt.Printf("查看日志: tail -f -n200 %s%s.$(date +%%F).log\n", OptionsLogDir, projectName)
	fmt.Printf("查看out: tail -f -n200 %s%s.out\n", OptionsLogDir, projectName)
	fmt.Printf("查看进程: ps -ef | grep %s\n", projectName)
}

func getProjectName() string {
	keyword := prompt.InputUi.SearchKeyword()

	if keyword != "" {
		fmt.Printf("正在匹配包含'%s'的项目\n", keyword)
	}

	dirFiles, err := ioutil.ReadDir(OptionsWorkDir)
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
		if prompt.SelectUi.IsAgree("未找到匹配的项目, 是否重新搜索?") {
			return getProjectName()
		}
		os.Exit(0)
	}

	if len(projects) == 1 {
		fmt.Printf("找到唯一匹配的项目: %s\n", projects[0])
		return projects[0]
	}

	return prompt.SelectUi.Project(projects)
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
	return prompt.SelectUi.Yml(ymlNames)
}

func getOldPidCmd(projectName, ymlName string) *exec.Cmd {
	psCmd, err := cmd.ExecInDir("", "ps", "-ef")
	if err != nil {
		return nil
	}

	grepCmd, err := cmd.ExecWithPreCmd(psCmd, "grep", OptionsWorkDir)
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

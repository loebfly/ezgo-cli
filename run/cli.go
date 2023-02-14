package run

import (
	"bytes"
	"ezgo-cli/cmd"
	"ezgo-cli/run/prompt"
	"ezgo-cli/tools"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var (
	// CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build
	OptionsWorkDir   = ""     // -workDir 项目根目录
	OptionsLogDir    = ""     // -logDir 日志目录
	OptionsSwagInit  = "y"    // -swagInit 是否初始化swagger
	OptionsGoBuild   = "y"    // -build 是否编译
	OptionsGoVersion = "1.17" // -goVersion 指定go版本
	OptionsGroup     = ""     // -group 指定项目组
	OptionsBatch     = "n"    // -batch 是否批量执行
)

func Exec() {
	cmdFlag := flag.NewFlagSet(cmd.Run, flag.ExitOnError)
	cmdFlag.StringVar(&OptionsWorkDir, "workDir", "/opt/go/src/flamecloud.cn/", "项目根目录")
	cmdFlag.StringVar(&OptionsLogDir, "logDir", "/opt/logs/", "日志根目录")
	cmdFlag.StringVar(&OptionsSwagInit, "swag", "y", "是否生成swagger文档")
	cmdFlag.StringVar(&OptionsGoBuild, "build", "y", "是否要编译项目")
	cmdFlag.StringVar(&OptionsBatch, "batch", "n", "是否要批量操作")
	err := cmdFlag.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("解析命令行参数失败: ", err.Error())
		return
	}

	OptionsGroup = prompt.SelectUi.ProjectGroup()
	if OptionsGroup != "root" {
		OptionsWorkDir = OptionsWorkDir + OptionsGroup + "/"
	}

	fmt.Println("项目根目录: ", OptionsWorkDir)
	fmt.Println("日志根目录: ", OptionsLogDir)

	if OptionsBatch == "y" {
		// 批量执行
		var projectNames = getAllProjectNameForWorkDir()
		for len(projectNames) > 0 {
			projectName := projectNames[0]
			if prompt.SelectUi.IsAgree("是否要执行(" + projectName + ")项目?") {
				startRunFlow(projectName)
			}
			projectNames = projectNames[1:]
		}
	} else {
		projectName := getProjectName()
		startRunFlow(projectName)
	}
}

func startRunFlow(projectName string) {
	projectDir := fmt.Sprintf("%s%s", OptionsWorkDir, projectName)

	_, err := os.Stat(projectDir + "/go.mod")
	if err == nil || os.IsExist(err) {
		// 读取go.mod文件
		content, err := tools.File(projectDir + "/go.mod").ReadString()
		if err != nil {
			fmt.Println("读取go.mod文件失败: ", err.Error())
			os.Exit(0)
		}
		OptionsGoVersion = "1.17"
		if strings.Contains(content, "go 1.17") {
			OptionsGoVersion = "1.17"
		} else if strings.Contains(content, "go 1.19") {
			OptionsGoVersion = "1.19"
		}
		fmt.Printf("读取到项目go版本: %s\n", OptionsGoVersion)
		// 设置go版本
		setGoVersion()
	} else {
		// 选择Go版本
		OptionsGoVersion = prompt.SelectUi.GoVersion()
		setGoVersion()
	}

	if OptionsGoBuild == "y" {
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

	if OptionsSwagInit == "y" {

		if OptionsGoVersion == "1.17" {
			fmt.Println("开始执行 swag init")
			_, err := cmd.ExecInDirWithPrint(projectDir, "swag", "init")
			if err != nil {
				fmt.Printf("生成swag文档失败: %s", err.Error())
			}
			fmt.Println("swag init 执行完毕")
		} else {
			_, err := cmd.ExecInDirWithPrint(projectDir, "swag", "init", "--pd", "--parseInternal")
			if err != nil {
				fmt.Printf("生成swag文档失败: %s", err.Error())
			}
		}

	}

	if OptionsGoBuild == "y" {
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

// getAllProjectNameForWorkDir 获取项目目录下的所有项目名称
func getAllProjectNameForWorkDir() []string {
	dirFiles, err := ioutil.ReadDir(OptionsWorkDir)
	if err != nil {
		fmt.Printf("读取项目目录失败: %s", err.Error())
		os.Exit(0)
	}

	var projects []string
	for _, dirFile := range dirFiles {
		if dirFile.IsDir() {
			projects = append(projects, dirFile.Name())
		}
	}

	return projects
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

func setGoVersion() {
	_, err := cmd.ExecInDir("", "rm", "-rf", "/usr/local/go")
	if err != nil {
		fmt.Printf("设置GO_VERSION失败: %s", err.Error())
		os.Exit(0)
		return
	}
	_, err = cmd.ExecInDir("", "ln", "-sf", "go"+OptionsGoVersion, "/usr/local/go")
	if err != nil {
		fmt.Printf("设置GO_VERSION失败: %s", err.Error())
		os.Exit(0)
		return
	}
	fmt.Printf("设置GO %s版本成功\n", OptionsGoVersion)
}

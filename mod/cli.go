package mod

import (
	"ezgo-cli/cmd"
	"ezgo-cli/mod/prompt"
	"ezgo-cli/tools"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	OptionsWorkDir  = "" // -workDir 项目根目录
	OptionsPackages = "" // 升级的包, 多个包用逗号分隔
)

func Exec() {
	if os.Args[2] == "-help" {
		printHelp()
		return
	}
	cmdFlag := flag.NewFlagSet(cmd.ModUpdate, flag.ExitOnError)
	cmdFlag.StringVar(&OptionsWorkDir, "dir", "", "项目根目录")
	cmdFlag.StringVar(&OptionsPackages, "pkg", "", "要升级的包(包含版本号), 例如: github.com/loebfly/ezgin@v0.1.36, 多个包用逗号分隔")
	err := cmdFlag.Parse(os.Args[3:])
	if err != nil {
		fmt.Println("解析命令行参数失败: ", err.Error())
		return
	}
	if OptionsWorkDir == "" {
		fmt.Println("项目根目录不能为空")
		return
	}
	if OptionsPackages == "" {
		fmt.Println("要升级的包不能为空")
		return
	}
	upgradePackageList := strings.Split(OptionsPackages, ",")
	if len(upgradePackageList) == 0 {
		fmt.Println("要升级的包不能为空")
		return
	}
	fmt.Println("项目根目录: ", OptionsWorkDir)
	fmt.Printf("要升级的包:\n%s\n", strings.Join(upgradePackageList, " \n"))

	// 找项目
	var projects = getProjects()

	// 找go.mod
	for _, project := range projects {

		goModPath := filepath.Join(OptionsWorkDir, project, "go.mod")
		_, err = os.Stat(goModPath)
		if err != nil {
			fmt.Printf("项目: %s, 没有go.mod文件, 跳过", project)
			continue
		}

		// 执行go mod
		isNeedTidy := false
		for _, upgradePackage := range upgradePackageList {
			packageInfo := strings.Split(upgradePackage, "@")
			if len(packageInfo) != 2 {
				fmt.Printf("包: %s, 格式错误, 跳过\n", upgradePackage)
				continue
			}
			packageName := packageInfo[0]
			fmt.Printf("项目: %s, 升级包: %s\n", project, upgradePackage)

			// 判断goModPath是否存在packageName
			goModContent, err := ioutil.ReadFile(goModPath)
			if err != nil {
				fmt.Printf("读取go.mod文件失败: %s\n", err.Error())
				continue
			}

			if !strings.Contains(string(goModContent), packageName) {
				fmt.Printf("go.mod文件中未找到%s包, 跳过\n", packageName)
				continue
			} else {
				//if !prompt.SelectUi.IsAgree(fmt.Sprintf("是否升级项目'%s'?", project)) {
				//	continue
				//}
				fmt.Println("go.mod文件中找到该包, 替换中...")
			}

			packageList := strings.Split(string(goModContent), "\n")
			for i, packageItem := range packageList {
				if packageItem == "require (" {
					continue
				}
				if strings.Contains(packageItem, packageName) {
					packageList[i] = fmt.Sprintf("%s %s", packageName, packageInfo[1])
					isNeedTidy = true
					break
				}
				if packageItem == ")" {
					break
				}
			}
			if isNeedTidy {
				err = tools.File(goModPath).WriteString(strings.Join(packageList, "\n"))
				if err != nil {
					fmt.Printf("写入go.mod文件失败: %s\n", err.Error())
					return
				}
				fmt.Println("go mod 替换该包成功")
			}
		}
		if isNeedTidy {
			fmt.Println("go mod tidy中...")
			projectDir := filepath.Join(OptionsWorkDir, project)
			goVersion := getProjectGoVersion(projectDir)
			setGoVersion(goVersion)
			_, err = cmd.ExecInDirWithPrint(projectDir, "go", "mod", "tidy", "-compat="+goVersion)
			if err != nil {
				fmt.Printf("执行go mod tidy失败: %s", err.Error())
				continue
			}
			fmt.Println("go mod tidy成功")
		}
	}
}

func getProjects() []string {
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
			return getProjects()
		}
		os.Exit(0)
	}
	fmt.Printf("找到以下项目:\n%s\n ", strings.Join(projects, "\n"))
	return projects
}

func getProjectGoVersion(projectDir string) string {
	// 读取go.mod文件
	content, err := tools.File(projectDir + "/go.mod").ReadString()
	if err != nil {
		fmt.Println("读取go.mod文件失败: ", err.Error())
		os.Exit(0)
	}
	goVersion := "1.17"
	if strings.Contains(content, "go 1.17") {
		goVersion = "1.17"
	} else if strings.Contains(content, "go 1.19") {
		goVersion = "1.19"
	}
	return goVersion
}

func setGoVersion(version string) {
	_, err := cmd.ExecInDir("", "rm", "-rf", "/usr/local/go")
	if err != nil {
		fmt.Printf("设置GO_VERSION失败: %s", err.Error())
		os.Exit(0)
		return
	}
	_, err = cmd.ExecInDir("", "ln", "-sf", "go"+version, "/usr/local/go")
	if err != nil {
		fmt.Printf("设置GO_VERSION失败: %s", err.Error())
		os.Exit(0)
		return
	}
	fmt.Printf("设置GO %s版本成功\n", version)
}

func printHelp() {
	fmt.Println("用法: ezgo-cli mod [options]")
	fmt.Println("options:")
	fmt.Println("  upgrade")
	fmt.Println("  ezgo-cli mod upgrade -help 查看用法")
}

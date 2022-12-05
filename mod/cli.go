package mod

import (
	"ezgo-cli/cmd"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var (
	OptionsWorkDir     = ""     // -workDir 项目根目录
	OptionsPackages    = ""     // 升级的包, 多个包用逗号分隔
	upgradePackageList []string // -upgradePackage 升级的包
)

func Exec() {
	cmdFlag := flag.NewFlagSet(cmd.ModUpdate, flag.ExitOnError)
	cmdFlag.StringVar(&OptionsWorkDir, "workDir", "", "项目根目录")
	cmdFlag.StringVar(&OptionsPackages, "packages", "", "要升级的包(包含版本号), 例如: github.com/loebfly/ezgin@v0.1.36, 多个包用逗号分隔")
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
	upgradePackageList = strings.Split(OptionsPackages, ",")
	if len(upgradePackageList) == 0 {
		fmt.Println("要升级的包不能为空")
		return
	}
	fmt.Println("项目根目录: ", OptionsWorkDir)
	fmt.Println("要升级的包: ", OptionsPackages)

	// 找项目
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

	if len(projects) == 0 {
		fmt.Println("项目根目录下没有项目")
		os.Exit(0)
	}

	// 找go.mod
	for _, project := range projects {
		goModPath := filepath.Join(OptionsWorkDir, project, "go.mod")
		_, err := os.Stat(goModPath)
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
				fmt.Printf("go.mod文件中未找到该包, 跳过\n")
				continue
			} else {
				fmt.Println("go.mod文件中找到该包, 替换中...")
			}

			// 执行go mod edit -require
			_, err = cmd.ExecInDirWithPrint(filepath.Join(OptionsWorkDir, project), "go", "mod", "edit", "-require", upgradePackage)
			if err != nil {
				fmt.Printf("执行go mod edit -require失败: %s", err.Error())
				continue
			}
			fmt.Println("go mod 替换该包成功")
			isNeedTidy = true
		}
		if isNeedTidy {
			fmt.Println("go mod tidy中...")
			_, err = cmd.ExecInDirWithPrint(filepath.Join(OptionsWorkDir, project), "go", "mod", "tidy", "-compat=1.17")
			if err != nil {
				fmt.Printf("执行go mod tidy失败: %s", err.Error())
				continue
			}
			fmt.Println("go mod tidy成功")
		}
	}
}

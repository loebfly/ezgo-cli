package new

import (
	"ezgo-cli/cmd"
	"ezgo-cli/new/ezgin"
	"ezgo-cli/tools"
	"flag"
	"fmt"
	"github.com/levigross/grequests"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	EZGinExampleGitUrl = "https://github.com/loebfly/ezgin-example/archive/refs/heads/main.zip"
)

var (
	ProjectDir string // 项目目录
)

func Exec() {
	cmdFlag := flag.NewFlagSet("new", flag.ExitOnError)
	cmdFlag.StringVar(&ProjectDir, "dir", "", "项目目录")
	err := cmdFlag.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("解析命令行参数失败: ", err.Error())
		return
	}
	if ProjectDir == "" {
		fmt.Println("项目目录不能为空")
		return
	}

	fmt.Println("开始制作项目")

	fmt.Println("项目目录: ", ProjectDir)
	projectName := filepath.Base(ProjectDir)
	fmt.Println("项目名称: ", projectName)

	fmt.Println("清空项目目录...")
	_ = tools.File(ProjectDir).DeleteDirSubFiles()
	fmt.Println("清空项目目录成功")

	fmt.Println("准备项目模板...")
	response, err := grequests.Get(EZGinExampleGitUrl, nil)
	zipFile := ProjectDir + "/ezgin-example.zip"
	if err := response.DownloadToFile(zipFile); err != nil {
		fmt.Println("项目模板下载失败: ", err.Error())
		return
	}
	fmt.Println("项目模板已准备")

	// 解压项目模板
	_ = tools.File(zipFile).UnzipTo(ProjectDir)

	// 删除压缩包
	_ = os.Remove(zipFile)

	// 移动文件到项目目录
	exampleDir := filepath.Join(ProjectDir, "ezgin-example-main")
	_ = tools.File(exampleDir).MoveDirSubFilesTo(ProjectDir)

	// 删除临时文件夹
	_ = os.RemoveAll(exampleDir)

	ezginCfg := GetEzginCfg()
	ezginCfg.App.Name = projectName
	WriteYml(ezginCfg)

	// 执行 go mod init 命令
	fmt.Println("开始执行 go mod init")
	_, err = cmd.ExecInDirWithPrint(ProjectDir, "go", "mod", "init", projectName)
	if err != nil {
		fmt.Println("go mod init 失败: ", err.Error())
		return
	}
	fmt.Println("go mod init 成功")

	fmt.Println("开始执行 go mod tidy")
	_, err = cmd.ExecInDirWithPrint(ProjectDir, "go", "mod", "tidy", "-compat=1.17")
	if err != nil {
		fmt.Println("go mod tidy 执行失败: ", err.Error())
		os.Exit(0)
	}
	fmt.Println("go mod tidy 执行完毕")
}

func GetEzginCfg() ezgin.Config {

	return ezgin.Config{}
}

func WriteYml(cfg ezgin.Config) {

	yml, _ := tools.File(ProjectDir + "/ezgin.yml").ReadString()

	yml = strings.ReplaceAll(yml, "{app-name}", cfg.App.Name)
	yml = strings.ReplaceAll(yml, "{app-ip}", cfg.App.Version)
	yml = strings.ReplaceAll(yml, "{app-port}", strconv.Itoa(cfg.App.Port))
	yml = strings.ReplaceAll(yml, "{app-port-ssl}", strconv.Itoa(cfg.App.PortSsl))
	yml = strings.ReplaceAll(yml, "{app-cert}", cfg.App.Cert)
	yml = strings.ReplaceAll(yml, "{app-key}", cfg.App.Key)
	yml = strings.ReplaceAll(yml, "{app-debug}", strconv.FormatBool(cfg.App.Debug))
	yml = strings.ReplaceAll(yml, "{app-version}", cfg.App.Version)
	yml = strings.ReplaceAll(yml, "{app-env}", cfg.App.Env)

	yml = strings.ReplaceAll(yml, "{nacos-server}", cfg.Nacos.Server)
	yml = strings.ReplaceAll(yml, "{nacos-yml-nacos}", cfg.Nacos.Nacos)
	yml = strings.ReplaceAll(yml, "{nacos-yml-mysql}", cfg.Nacos.Mysql)
	yml = strings.ReplaceAll(yml, "{nacos-yml-mongo}", cfg.Nacos.Mongo)
	yml = strings.ReplaceAll(yml, "{nacos-yml-redis}", cfg.Nacos.Redis)
	yml = strings.ReplaceAll(yml, "{nacos-yml-kafka}", cfg.Nacos.Kafka)

	yml = strings.ReplaceAll(yml, "{gin-mode}", cfg.Gin.Mode)
	yml = strings.ReplaceAll(yml, "{gin-middleware}", cfg.Gin.Middleware)
	yml = strings.ReplaceAll(yml, "{gin-mw_logs-mongo_tag}", cfg.Gin.MongoTag)
	yml = strings.ReplaceAll(yml, "{gin-mw_logs-mongo_table}", cfg.Gin.MongoTable)
	yml = strings.ReplaceAll(yml, "{gin-kafka_topic}", cfg.Gin.KafkaTopic)

	yml = strings.ReplaceAll(yml, "{logs-level}", cfg.Logs.Level)
	yml = strings.ReplaceAll(yml, "{logs-out}", cfg.Logs.Out)
	yml = strings.ReplaceAll(yml, "{logs-file}", cfg.Logs.File)

	yml = strings.ReplaceAll(yml, "{i18n-app_name}", cfg.I18n.AppName)
	yml = strings.ReplaceAll(yml, "{i18n-server_name}", cfg.I18n.ServerName)
	yml = strings.ReplaceAll(yml, "{i18n-check_uri}", cfg.I18n.CheckUri)
	yml = strings.ReplaceAll(yml, "{i18n-query_uri}", cfg.I18n.QueryUri)
	yml = strings.ReplaceAll(yml, "{i18n-duration}", strconv.Itoa(cfg.I18n.Duration))

	_ = tools.File(ProjectDir + "/ezgin.yml").WriteString(yml)

	// 重命名为cfg.App.Name
	_ = os.Rename(ProjectDir+"/ezgin.yml", ProjectDir+"/"+cfg.App.Name+".yml")
	fmt.Println(yml)
}

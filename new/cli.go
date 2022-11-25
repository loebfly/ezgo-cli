package new

import (
	"ezgo-cli/cmd"
	"ezgo-cli/new/ezgin"
	"ezgo-cli/new/prompt"
	"ezgo-cli/tools"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"path/filepath"
	"strings"
)

const (
	exampleGit         = "https://gitee.com/loebfly/ezgin-example.git"
	exampleYml         = "ezgin.yml"
	exampleProjectName = "ezgin-example"
)

var (
	ProjectDir   string // 项目目录
	ProjectGroup string // 项目分组
	UseTemplate  bool   // 项目模板
)

func Exec() {
	cmdFlag := flag.NewFlagSet("new", flag.ExitOnError)
	cmdFlag.StringVar(&ProjectDir, "dir", "", "项目目录, 不可为当前目录")
	cmdFlag.StringVar(&ProjectGroup, "group", "", "项目分组")
	cmdFlag.BoolVar(&UseTemplate, "ue", false, "生成的项目是否带模块示例")
	err := cmdFlag.Parse(os.Args[2:])
	if err != nil {
		fmt.Println("解析命令行参数失败: ", err.Error())
		return
	}
	if ProjectDir == "" {
		fmt.Println("项目目录不能为空")
		return
	}

	// 检查目录是否与当前目录相同
	absPath, _ := filepath.Abs(os.Args[0])
	if absPath == ProjectDir {
		fmt.Println("项目目录不能为当前目录")
		return
	}

	fmt.Println("开始制作项目")

	fmt.Println("项目目录: ", ProjectDir)
	projectName := filepath.Base(ProjectDir)
	fmt.Println("项目名称: ", projectName)

	_, err = os.Stat(ProjectDir)
	if err != nil {
		// 不存在则创建
		err = os.MkdirAll(ProjectDir, os.ModePerm)
		if err != nil {
			fmt.Println("创建项目目录失败: ", err.Error())
			return
		}
	} else {
		fmt.Println("清空项目目录...")
		err = tools.File(ProjectDir).DeleteDirSubFiles()
		if err != nil {
			fmt.Println("清空项目目录失败: ", err.Error())
		}
		fmt.Println("清空项目目录成功")
	}

	isUseDefaultYml := prompt.InputUi.Run(promptExit, promptui.Prompt{
		Label:     "是否生成默认程序yml配置?",
		IsConfirm: true,
		Default:   "Y",
	})
	ezCfg := ezgin.Config{
		UseTemplate: UseTemplate,
	}
	if isUseDefaultYml == "Y" || isUseDefaultYml == "y" {
		ezCfg = ezgin.GetDefaultConfig(ProjectGroup, projectName)
	} else {
		ezCfg = ezgin.GetCustomConfig(promptExit, projectName)
	}

	fmt.Println("正在准备项目模板...")
	_, err = cmd.ExecInDir(ProjectDir, "git", "clone", exampleGit)
	if err != nil {
		fmt.Println("准备项目模板失败: ", err.Error())
		return
	}

	// 移动文件到项目目录
	exampleDir := filepath.Join(ProjectDir, exampleProjectName)
	_ = tools.File(exampleDir).MoveDirSubShowFilesTo(ProjectDir)

	// 删除临时文件夹
	_ = os.RemoveAll(exampleDir)

	fmt.Println("项目模板已准备")

	fmt.Println("开始生成项目配置文件...")
	WriteYml(ezCfg)
	fmt.Println("项目配置文件已生成")

	fmt.Println("正在修复项目引用...")
	fixProject(ezCfg)
	fmt.Println("项目引用修复完成")

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

	fmt.Println("开始执行 swag init")
	_, err = cmd.ExecInDirWithPrint(ProjectDir, "swag", "init")
	fmt.Println("swag init 执行完毕")

	fmt.Println("项目制作完毕")
	os.Exit(0)
}

func WriteYml(cfg ezgin.Config) {
	ymlPath := filepath.Join(ProjectDir, exampleYml)
	yml, _ := tools.File(ymlPath).ReadString()

	yml = strings.ReplaceAll(yml, "{app-name}", cfg.App.Name)
	if cfg.App.Ip == "" {
		// 删除 ip 配置
		yml = strings.ReplaceAll(yml, "ip: {app-ip}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{app-ip}", cfg.App.Ip)
	}
	if cfg.App.Port == "" || cfg.App.Port == "0" {
		// 删除 port 配置
		yml = strings.ReplaceAll(yml, "port: {app-port}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{app-port}", cfg.App.Port)
	}
	if cfg.App.PortSsl == "" || cfg.App.PortSsl == "0" {
		// 删除 port-ssl 配置
		yml = strings.ReplaceAll(yml, "port-ssl: {app-port-ssl}", "")
		yml = strings.ReplaceAll(yml, "cert: {app-cert}", "")
		yml = strings.ReplaceAll(yml, "key: {app-key}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{app-port-ssl}", cfg.App.PortSsl)
		yml = strings.ReplaceAll(yml, "{app-cert}", cfg.App.Cert)
		yml = strings.ReplaceAll(yml, "{app-key}", cfg.App.Key)
	}

	yml = strings.ReplaceAll(yml, "{app-debug}", cfg.App.Debug)
	yml = strings.ReplaceAll(yml, "{app-version}", cfg.App.Version)
	yml = strings.ReplaceAll(yml, "{app-env}", cfg.App.Env)

	if cfg.Nacos.Server == "" {
		// 删除 nacos 配置
		yml = strings.ReplaceAll(yml, "nacos:", "")
		yml = strings.ReplaceAll(yml, "server: {nacos-server}", "")
		yml = strings.ReplaceAll(yml, "yml:", "")
		yml = strings.ReplaceAll(yml, "nacos: {nacos-yml-nacos}", "")
		yml = strings.ReplaceAll(yml, "mysql: {nacos-yml-mysql}", "")
		yml = strings.ReplaceAll(yml, "mongo: {nacos-yml-mongo}", "")
		yml = strings.ReplaceAll(yml, "redis: {nacos-yml-redis}", "")
		yml = strings.ReplaceAll(yml, "kafka: {nacos-yml-kafka}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{nacos-server}", cfg.Nacos.Server)
		yml = strings.ReplaceAll(yml, "{nacos-yml-nacos}", cfg.Nacos.Nacos)
		yml = strings.ReplaceAll(yml, "{nacos-yml-mysql}", cfg.Nacos.Mysql)
		yml = strings.ReplaceAll(yml, "{nacos-yml-mongo}", cfg.Nacos.Mongo)
		yml = strings.ReplaceAll(yml, "{nacos-yml-redis}", cfg.Nacos.Redis)
		yml = strings.ReplaceAll(yml, "{nacos-yml-kafka}", cfg.Nacos.Kafka)
	}

	yml = strings.ReplaceAll(yml, "{gin-mode}", cfg.Gin.Mode)
	if cfg.Gin.Middleware == "-" {
		yml = strings.ReplaceAll(yml, "{gin-middleware}", "\"-\"")
	} else if cfg.Gin.Middleware == "" {
		// 删除 middleware 配置
		yml = strings.ReplaceAll(yml, "middleware: {gin-middleware}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{gin-middleware}", cfg.Gin.Middleware)
	}
	if cfg.Gin.MongoTag == "-" {
		yml = strings.ReplaceAll(yml, "{gin-mw_logs-mongo_tag}", "\"-\"")
	} else if cfg.Gin.MongoTag == "" {
		// 删除 mongo-tag 配置
		yml = strings.ReplaceAll(yml, "mongo_tag: {gin-mw_logs-mongo_tag}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{gin-mw_logs-mongo_tag}", cfg.Gin.MongoTag)
	}

	if cfg.Gin.MongoTable == "" {
		yml = strings.ReplaceAll(yml, "mongo_table: {gin-mw_logs-mongo_table}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{gin-mw_logs-mongo_table}", cfg.Gin.MongoTable)
	}

	if cfg.Gin.KafkaTopic == "-" {
		yml = strings.ReplaceAll(yml, "{gin-mw_logs-kafka_topic}", "\"-\"")
	} else {
		yml = strings.ReplaceAll(yml, "{gin-mw_logs-kafka_topic}", cfg.Gin.KafkaTopic)
	}

	if cfg.Logs.Level == "-" {
		yml = strings.ReplaceAll(yml, "{logs-level}", "\"-\"")
		yml = strings.ReplaceAll(yml, "out: {logs-out}", "")
		yml = strings.ReplaceAll(yml, "file: {logs-file}", "")
	} else {
		yml = strings.ReplaceAll(yml, "{logs-level}", cfg.Logs.Level)
		yml = strings.ReplaceAll(yml, "{logs-out}", cfg.Logs.Out)
		if cfg.Logs.File == "" {
			// 删除 file 配置
			yml = strings.ReplaceAll(yml, "file: {logs-file}", "")
		} else {
			yml = strings.ReplaceAll(yml, "{logs-file}", cfg.Logs.File)
		}
	}

	if cfg.I18n.AppName == "-" {
		yml = strings.ReplaceAll(yml, "{i18n-app_name}", "\"-\"")
		yml = strings.ReplaceAll(yml, "server_name: {i18n-server_name}", "")
		yml = strings.ReplaceAll(yml, "check_uri: {i18n-check_uri}", "")
		yml = strings.ReplaceAll(yml, "query_uri: {i18n-query_uri}", "")
		yml = strings.ReplaceAll(yml, "duration: {i18n-duration}", "")
	} else {
		if cfg.I18n.AppName == "" {
			// 删除 i18n 配置
			yml = strings.ReplaceAll(yml, "i18n:", "")
			yml = strings.ReplaceAll(yml, "app_name: {i18n-app_name}", "")
			yml = strings.ReplaceAll(yml, "server_name: {i18n-server_name}", "")
			yml = strings.ReplaceAll(yml, "check_uri: {i18n-check_uri}", "")
			yml = strings.ReplaceAll(yml, "query_uri: {i18n-query_uri}", "")
			yml = strings.ReplaceAll(yml, "duration: {i18n-duration}", "")
		} else {
			yml = strings.ReplaceAll(yml, "{i18n-app_name}", cfg.I18n.AppName)
			if cfg.I18n.ServerName == "" {
				// 删除 server_name 配置
				yml = strings.ReplaceAll(yml, "server_name: {i18n-server_name}", "")
			} else {
				yml = strings.ReplaceAll(yml, "{i18n-server_name}", cfg.I18n.ServerName)
			}
			if cfg.I18n.CheckUri == "" {
				// 删除 check_uri 配置
				yml = strings.ReplaceAll(yml, "check_uri: {i18n-check_uri}", "")
			} else {
				yml = strings.ReplaceAll(yml, "{i18n-check_uri}", cfg.I18n.CheckUri)
			}
			if cfg.I18n.QueryUri == "" {
				// 删除 query_uri 配置
				yml = strings.ReplaceAll(yml, "query_uri: {i18n-query_uri}", "")
			} else {
				yml = strings.ReplaceAll(yml, "{i18n-query_uri}", cfg.I18n.QueryUri)
			}
			if cfg.I18n.Duration == "" {
				// 删除 duration 配置
				yml = strings.ReplaceAll(yml, "duration: {i18n-duration}", "")
			} else {
				yml = strings.ReplaceAll(yml, "{i18n-duration}", cfg.I18n.Duration)
			}
		}
	}

	items := strings.Split(yml, "\n")
	var newItems []string
	for _, item := range items {
		if strings.TrimSpace(item) == "" {
			continue
		}
		newItems = append(newItems, item)
	}

	_ = tools.File(ymlPath).WriteString(strings.Join(newItems, "\n"))

	// 重命名为cfg.App.Name
	newYmlPath := filepath.Join(ProjectDir, cfg.App.Name+".yml")
	_ = os.Rename(ymlPath, newYmlPath)
}

func fixProject(cfg ezgin.Config) {
	// 修改main.go
	mainPath := filepath.Join(ProjectDir, "main.go")
	main, _ := tools.File(mainPath).ReadString()
	main = strings.ReplaceAll(main, "{swag-title}", cfg.App.Name)
	main = strings.ReplaceAll(main, "{swag-version}", cfg.App.Version)
	main = strings.ReplaceAll(main, "{swag-description}", "基于ezgin框架，由ezgo-cli生成")
	main = strings.ReplaceAll(main, exampleProjectName, cfg.App.Name)
	_ = tools.File(mainPath).WriteString(main)

	controllerTemplatePath := filepath.Join(ProjectDir, "controller", "template.go")
	if cfg.UseTemplate {
		// 修改controller/template.go
		controllerTemplate, _ := tools.File(controllerTemplatePath).ReadString()
		controllerTemplate = strings.ReplaceAll(controllerTemplate, exampleProjectName, cfg.App.Name)
		_ = tools.File(controllerTemplatePath).WriteString(controllerTemplate)
	} else {
		_ = os.Remove(controllerTemplatePath)
		enterPath := filepath.Join(ProjectDir, "controller", "enter.go")
		enter, _ := tools.File(enterPath).ReadString()
		enter = strings.ReplaceAll(enter, "var Template = new(templateController)", "")
		_ = tools.File(enterPath).WriteString(enter)
	}

	if cfg.Nacos.Mongo == "" {
		_ = os.RemoveAll(filepath.Join(ProjectDir, "mongo"))
	} else {
		mongoTemplatePath := filepath.Join(ProjectDir, "mongo", "template.go")
		if cfg.UseTemplate {
			// 修改mongo/template.go
			mongoTemplate, _ := tools.File(mongoTemplatePath).ReadString()
			mongoTemplate = strings.ReplaceAll(mongoTemplate, exampleProjectName, cfg.App.Name)
			_ = tools.File(mongoTemplatePath).WriteString(mongoTemplate)
		} else {
			_ = os.Remove(mongoTemplatePath)
			enterPath := filepath.Join(ProjectDir, "mongo", "enter.go")
			enter, _ := tools.File(enterPath).ReadString()
			enter = strings.ReplaceAll(enter, "var Template = new(templateMgo)", "")
			_ = tools.File(enterPath).WriteString(enter)
		}
	}

	if cfg.Nacos.Mysql == "" {
		_ = os.RemoveAll(filepath.Join(ProjectDir, "mysql"))
	} else {
		mysqlTemplatePath := filepath.Join(ProjectDir, "mysql", "template.go")
		if cfg.UseTemplate {
			// 修改mysql/template.go
			mysqlTemplate, _ := tools.File(mysqlTemplatePath).ReadString()
			mysqlTemplate = strings.ReplaceAll(mysqlTemplate, exampleProjectName, cfg.App.Name)
			_ = tools.File(mysqlTemplatePath).WriteString(mysqlTemplate)
		} else {
			_ = os.Remove(mysqlTemplatePath)
			enterPath := filepath.Join(ProjectDir, "mysql", "enter.go")
			enter, _ := tools.File(enterPath).ReadString()
			enter = strings.ReplaceAll(enter, "var Template = new(templateDao)", "")
			_ = tools.File(enterPath).WriteString(enter)
		}
	}

	if cfg.Nacos.Nacos == "" {
		_ = os.RemoveAll(filepath.Join(ProjectDir, "nacos"))
	} else {
		nacosTemplatePath := filepath.Join(ProjectDir, "nacos", "template.go")
		if cfg.UseTemplate {
			// 修改nacos/template.go
			nacosTemplate, _ := tools.File(nacosTemplatePath).ReadString()
			nacosTemplate = strings.ReplaceAll(nacosTemplate, exampleProjectName, cfg.App.Name)
			_ = tools.File(nacosTemplatePath).WriteString(nacosTemplate)
		} else {
			_ = os.Remove(nacosTemplatePath)
			enterPath := filepath.Join(ProjectDir, "nacos", "enter.go")
			enter, _ := tools.File(enterPath).ReadString()
			enter = strings.ReplaceAll(enter, "var Template = new(templateNacos)", "")
			_ = tools.File(enterPath).WriteString(enter)
		}
	}

	if cfg.Nacos.Redis == "" {
		_ = os.RemoveAll(filepath.Join(ProjectDir, "redis"))
	} else {
		redisTemplatePath := filepath.Join(ProjectDir, "redis", "template.go")
		if cfg.UseTemplate {
			// 修改redis/template.go
			redisTemplate, _ := tools.File(redisTemplatePath).ReadString()
			redisTemplate = strings.ReplaceAll(redisTemplate, exampleProjectName, cfg.App.Name)
			_ = tools.File(redisTemplatePath).WriteString(redisTemplate)
		} else {
			_ = os.Remove(redisTemplatePath)
			enterPath := filepath.Join(ProjectDir, "redis", "enter.go")
			enter, _ := tools.File(enterPath).ReadString()
			enter = strings.ReplaceAll(enter, "var Template = new(templateRds)", "")
			_ = tools.File(enterPath).WriteString(enter)
		}
	}
	// 修改router/router.go
	routerPath := filepath.Join(ProjectDir, "router", "router.go")
	router, _ := tools.File(routerPath).ReadString()
	if cfg.UseTemplate {
		router = strings.ReplaceAll(router, exampleProjectName, cfg.App.Name)
	} else {
		router = strings.ReplaceAll(router, "\n\t\"ezgin-example/controller\"", "")
		router = strings.ReplaceAll(router, "\n\t\t\"get\": controller.Template.Get,", "")
	}
	_ = tools.File(routerPath).WriteString(router)

	// 修改service/template.go
	serviceTemplatePath := filepath.Join(ProjectDir, "service", "template.go")
	serviceTemplate, _ := tools.File(serviceTemplatePath).ReadString()
	if cfg.UseTemplate {
		serviceTemplate = strings.ReplaceAll(serviceTemplate, exampleProjectName, cfg.App.Name)
	} else {
		_ = os.Remove(controllerTemplatePath)
		enterPath := filepath.Join(ProjectDir, "service", "enter.go")
		enter, _ := tools.File(enterPath).ReadString()
		enter = strings.ReplaceAll(enter, "var Template = new(templateService)", "")
		_ = tools.File(enterPath).WriteString(enter)
	}
	_ = tools.File(serviceTemplatePath).WriteString(serviceTemplate)
}

func promptExit() {
	_ = tools.File(ProjectDir).DeleteDirSubFiles()
}

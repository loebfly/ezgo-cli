package ezgin

import (
	"ezgo-cli/new/prompt"
	"fmt"
	"github.com/manifoldco/promptui"
	"strconv"
	"strings"
)

type Config struct {
	App         App
	Nacos       Nacos
	Gin         Gin
	Logs        Logs
	I18n        I18n
	UseTemplate bool
}

type App struct {
	Name    string // 应用名称
	Ip      string // 应用ip地址, 默认为本机ip
	Port    string // http服务端口
	PortSsl string // https服务端口
	Cert    string // 应用证书文件路径, 用于https, 如果不需要https,则不需要配置
	Key     string // 应用私钥文件路径, 用于https,如果不需要https,则不需要配置
	Debug   string // 是否开启debug模式, 默认false, 如果开启, 则不会被其他服务调用
	Version string // 版本号
	Env     string // 环境 test, dev, prod
} // 应用配置

type Nacos struct {
	Server string // nacos服务地址
	Nacos  string // nacos配置文件名 只需要配置文件的前缀，内部会自动拼接-$Env.yml, 如果不需要nacos配置文件,则不需要配置
	Mysql  string // mysql配置文件名 只需要配置文件的前缀，内部会自动拼接-$Env.yml, 多个配置文件用逗号分隔, 如果不需要mysql配置文件,则不需要配置
	Mongo  string // mongo配置文件名 只需要配置文件的前缀，内部会自动拼接-$Env.yml, 多个配置文件用逗号分隔, 如果不需要mongo配置文件,则不需要配置
	Redis  string // redis配置文件名 只需要配置文件的前缀，内部会自动拼接-$Env.yml, 多个配置文件用逗号分隔, 如果不需要redis配置文件,则不需要配置
	Kafka  string // kafka配置文件名 只需要配置文件的前缀，内部会自动拼接-$Env.yml, 只支持单个配置文件, 如果不需要kafka配置文件,则不需要配置
} // nacos配置

type Gin struct {
	Mode       string // gin模式 debug, release
	Middleware string // gin中间件, 用逗号分隔, 暂时支持cors,trace,logs,xlang,recover不填则默认全部开启, - 表示不开启
	MongoTag   string // 需要与Nacos.Yml.Mongo中配置文件名对应, 默认为Nacos.Yml.Mongo中第一个配置文件, - 表示不开启
	MongoTable string // 日志表名, 默认为${App.Name}APIRequestLogs
	KafkaTopic string // # kafka 消息主题, 默认为${App.Name}, 多个主题用逗号分隔, - 表示不开启
} // gin配置

type Logs struct {
	Level string // 日志级别 debug > info > warn > error, 默认debug即全部打印, - 表示不开启
	Out   string // 日志输出方式, 可选值: console, file 默认 console
	File  string // 日志文件路径, 如果Out包含file, 不填默认/opt/logs/${App.Name}, 生成的文件会带上.$(Date +%F).log
} // 日志配置

type I18n struct {
	AppName    string // i18n应用名称, 多个用逗号分隔, 默认为default,${App.Name}, - 表示不开启
	ServerName string // i18n微服务名称, 默认x-lang
	CheckUri   string // i18n服务检查uri, 默认/lang/string/app/version
	QueryUri   string // i18n服务查询uri, 默认/lang/string/list
	Duration   string //  i18n服务查询间隔, 默认360s
} // i18n配置

func GetDefaultConfig(groupName, projectName string) Config {
	if groupName == "" {
		groupName = projectName
	}
	return Config{
		App: App{
			Name:    projectName,
			Port:    "8080",
			Debug:   "true",
			Version: "1.0.0",
			Env:     "test",
		},
		Nacos: Nacos{
			Server: "http://59.56.77.23:58848/",
			Nacos:  "nacos-" + groupName + "-ezgin",
			Mongo:  "nacos-" + groupName + "-ezgin",
		},
		Gin: Gin{
			Mode:       "debug",
			KafkaTopic: "-",
		},
		Logs: Logs{
			Level: "debug",
			Out:   "console,file",
		},
		I18n: I18n{
			AppName: "default," + projectName,
		},
	}
}

func GetCustomConfig(exitFunc func(), projectName string) Config {
	fmt.Printf("准备 %s 的配置参数...\n", projectName)
	var cfg = Config{}
	cfg.App.Name = projectName

	fmt.Println("配置App相关参数...")
	cfg.App.Ip = prompt.InputUi.RunWithLabel(exitFunc, "请输入项目IP地址")

	isNeedPort := prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:     "是否需要配置端口号",
		IsConfirm: true,
		Default:   "y",
	})
	if isNeedPort == "y" {
		cfg.App.Port = prompt.InputUi.Run(exitFunc, promptui.Prompt{
			Label:   "请输入http端口号",
			Default: "8080",
			Validate: func(input string) error {
				port := strings.TrimSpace(input)
				if port != "" {
					if _, err := strconv.Atoi(port); err != nil {
						return fmt.Errorf("http端口号必须为数字")
					}
				}
				return nil
			},
		}, true)

		cfg.App.PortSsl = prompt.InputUi.Run(exitFunc, promptui.Prompt{
			Label:   "请输入https端口号",
			Default: "8443",
			Validate: func(input string) error {
				port := strings.TrimSpace(input)
				if port != "" {
					if _, err := strconv.Atoi(port); err != nil {
						return fmt.Errorf("https端口号必须为数字")
					}
				}
				return nil
			},
		}, true)

		cfg.App.Cert = prompt.InputUi.RunWithLabel(exitFunc, "请输入证书文件路径")
		cfg.App.Key = prompt.InputUi.RunWithLabel(exitFunc, "请输入私钥文件路径")
	}

	cfg.App.Debug = prompt.SelectUi.Run(exitFunc, promptui.Select{
		Label: "请选择是否开启debug模式",
		Items: []string{"true", "false"},
	})
	cfg.App.Version = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入项目版本号",
		Default: "1.0.0",
		Validate: func(input string) error {
			if strings.TrimSpace(input) == "" {
				return fmt.Errorf("项目版本号不能为空")
			}
			return nil
		},
	}, true)
	cfg.App.Env = prompt.SelectUi.Run(exitFunc, promptui.Select{
		Label: "请选择项目运行环境",
		Items: []string{"test", "dev", "prod"},
	})

	fmt.Println("配置Nacos相关参数...")
	fmt.Println("如不使用直接回车即可")
	cfg.Nacos.Server = prompt.InputUi.RunWithLabel(exitFunc, "请输入nacos服务地址")
	if cfg.Nacos.Server != "" {
		fmt.Println("配置在nacos注册中心的yml文件相关参数...")
		fmt.Println("1. 只需要配置文件的前缀，内部会自动拼接-$Env.yml")
		fmt.Println("2. 如果有多个配置，用逗号分隔(其中nacos、kafka不支持多个配置)")
		fmt.Println("3. 如不使用直接回车即可")
		cfg.Nacos.Nacos = prompt.InputUi.RunWithLabel(exitFunc, "请输入nacos的yml配置名称")
		cfg.Nacos.Mysql = prompt.InputUi.RunWithLabel(exitFunc, "请输入mysql的yml配置名称")
		cfg.Nacos.Mongo = prompt.InputUi.RunWithLabel(exitFunc, "请输入mongo的yml配置名称")
		cfg.Nacos.Redis = prompt.InputUi.RunWithLabel(exitFunc, "请输入redis的yml配置名称")
		cfg.Nacos.Kafka = prompt.InputUi.RunWithLabel(exitFunc, "请输入kafka的yml配置名称")
	}

	fmt.Println("配置gin相关参数...")
	cfg.Gin.Mode = prompt.SelectUi.Run(exitFunc, promptui.Select{
		Label: "请选择运行模式",
		Items: []string{"debug", "release"},
	})
	fmt.Println("配置中间件...")
	fmt.Println("说明:")
	fmt.Println("1. 请使用逗号分隔多个中间件")
	fmt.Println("2. 为空默认使用所有中间件")
	fmt.Println("3. 可选中间件: cors,trace,logs,xlang,recover")
	cfg.Gin.Middleware = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "cors,trace,logs,xlang,recover",
	})

	if cfg.Gin.Middleware == "" || strings.Contains(cfg.Gin.Middleware, "logs") {
		fmt.Println("配置日志中间件参数...")
		fmt.Println("1、mongo配置...")
		fmt.Println("配置说明:")
		fmt.Println("1. 需要与Nacos.Yml.Mongo中配置文件名中的tag一致, 默认为Nacos.Yml.Mongo中第一个配置文件的tag, - 表示不开启")
		fmt.Println("2. mongo中的表名, 默认为${App.Name}APIRequestLogs")
		cfg.Gin.MongoTag = prompt.InputUi.RunWithLabel(exitFunc, "请输入mongo的tag")
		cfg.Gin.MongoTable = prompt.InputUi.RunWithLabel(exitFunc, "请输入mongo的表名")
		fmt.Println("2、kafka配置...")
		fmt.Println("配置说明:")
		fmt.Println("默认为${App.Name}, 多个主题用逗号分隔, - 表示不开启")
		cfg.Gin.KafkaTopic = prompt.InputUi.RunWithLabel(exitFunc, "请输入kafka 消息主题")
	}

	fmt.Println("配置程序日志参数...")
	fmt.Println("1. 日志级别, debug > info > warn > error, 默认debug即全部输出, - 表示不开启")
	cfg.Logs.Level = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "debug",
	})
	fmt.Println("2. 日志输出方式, 可选值: console, file 默认 console")
	cfg.Logs.Out = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "console",
	})
	if strings.Contains(cfg.Logs.Out, "file") {
		fmt.Printf("3. 日志文件路径, 如果Out包含file, 不填默认/opt/logs/${App.Name}, 生成的文件会带上.$(Date +%%F).log\n")
		cfg.Logs.File = prompt.InputUi.Run(exitFunc, promptui.Prompt{
			Label: "请输入",
		})
	}

	fmt.Println("配置国际化参数...")
	fmt.Println("1. i18n应用名称, 多个用逗号分隔, 默认为default,${App.Name}, - 表示不开启")
	cfg.I18n.AppName = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "default," + cfg.App.Name,
	})
	fmt.Println("2. i18n微服务名称, 默认x-lang")
	cfg.I18n.ServerName = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "x-lang",
	})
	fmt.Println("3. i18n服务检查uri, 默认/lang/string/app/version")
	cfg.I18n.CheckUri = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "/lang/string/app/version",
	})
	fmt.Println("4. i18n服务查询uri, 默认/lang/string/list")
	cfg.I18n.QueryUri = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "/lang/string/list",
	})
	fmt.Println("5. i18n服务查询间隔, 默认360, 单位秒")
	cfg.I18n.Duration = prompt.InputUi.Run(exitFunc, promptui.Prompt{
		Label:   "请输入",
		Default: "360",
		Validate: func(input string) error {
			duration := strings.TrimSpace(input)
			if duration != "" {
				if _, err := strconv.Atoi(duration); err != nil {
					return fmt.Errorf("duration必须为数字")
				}
			}
			return nil
		},
	}, true)

	fmt.Printf("%s 的配置已准备\n", projectName)
	return cfg
}

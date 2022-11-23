package config

import (
	"fmt"
	"os"
	"path/filepath"

	"ezgo-cli/let"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/yaml"
	kFile "github.com/knadh/koanf/providers/file"
)

var (
	YmlData *koanf.Koanf
)

func Init() {
	path, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	ymlPath := filepath.Join(path, let.ConfigYmlName)
	// 判断配置文件是否存在
	_, err := os.Stat(ymlPath)
	if err != nil && !os.IsExist(err) {
		// 不存在
		fmt.Printf("请先在程序同级目录创建名称为%s的配置文件\n", let.ConfigYmlName)
		fmt.Println("配置文件示例:")
		template := `
ezgo-cli:
   run:
     project_dir: # 项目根目录
     log_dir: 日志文件根目录
		`
		fmt.Println(template)
		os.Exit(1)
	}
	YmlData = koanf.New(".")
	f := kFile.Provider(ymlPath)
	err = YmlData.Load(f, yaml.Parser())
	if err != nil {
		fmt.Printf("配置文件解析错误:%s", err.Error())
		os.Exit(1)
	}
}

func GetProjectDir() string {
	val := YmlData.String(let.ProjectDirKey)
	if val == "" {
		val = "/ezgo-cli/project/"
	}
	return val
}

func GetLogDir() string {
	val := YmlData.String(let.LogDirKey)
	if val == "" {
		val = "/ezgo-cli/logs/"
	}
	return val
}

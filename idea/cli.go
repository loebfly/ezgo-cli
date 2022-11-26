package idea

import (
	"ezgo-cli/cmd"
	"ezgo-cli/tools"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	LocalDir  string // 项目本地根目录
	SftpName  string // sftp名称
	ServerDir string // 项目服务器根目录
)

func Exec() {
	switch os.Args[2] {
	case cmd.IdeaDeploy:
		ideaDevelopment()
	}
}

func ideaDevelopment() {
	cmdFlag := flag.NewFlagSet(cmd.IdeaDeploy, flag.ExitOnError)
	cmdFlag.StringVar(&LocalDir, "localDir", "", "项目本地目录, 项目自身目录或项目上级目录, 如果是上级目录, 会批量配置所有项目")
	cmdFlag.StringVar(&SftpName, "sftpName", "", "sftp名称, 多个用逗号分隔")
	cmdFlag.StringVar(&ServerDir, "serverDir", "", "项目服务器映射根目录")
	err := cmdFlag.Parse(os.Args[3:])
	if err != nil {
		fmt.Println("解析命令行参数失败: ", err.Error())
		return
	}

	if LocalDir == "" {
		fmt.Println("localDir参数不能为空")
		return
	}

	if SftpName == "" {
		fmt.Println("sftpName参数不能为空")
		return
	}

	if ServerDir == "" {
		fmt.Println("serverDir参数不能为空")
		return
	}

	_, err = os.Stat(filepath.Join(LocalDir, "go.mod"))
	if err == nil {
		projectName := path.Base(LocalDir)
		fmt.Println("正在处理: ", projectName)
		ideaPath := filepath.Join(LocalDir, ".idea")
		if tools.File(ideaPath).Exists() == false {
			fmt.Println("未找到.idea目录, 跳过")
			return
		}
		deployPath := path.Join(ideaPath, "deployment.xml")
		fmt.Println("部署文件路径: ", deployPath)
		if tools.File(deployPath).Exists() {
			fmt.Println("部署文件已存在, 跳过")
			return
		}
		fmt.Println("部署文件不存在, 创建中...")
		err = tools.File(deployPath).WriteString(getDeploymentXml(projectName))
		if err != nil {
			fmt.Println("写入部署xml失败: ", err.Error())
			return
		}
		fmt.Println("写入部署xml成功")
		return
	}

	fis, err := ioutil.ReadDir(LocalDir)
	if err != nil {
		fmt.Println("读取项目本地目录失败: ", err.Error())
		return
	}
	for _, file := range fis {
		fmt.Println("正在处理: ", file.Name())
		ideaPath := filepath.Join(LocalDir, file.Name(), ".idea")
		if tools.File(ideaPath).Exists() == false {
			fmt.Println("未找到.idea目录, 跳过")
			continue
		}
		deployPath := path.Join(ideaPath, "deployment.xml")
		fmt.Println("部署文件路径: ", deployPath)
		if tools.File(deployPath).Exists() {
			fmt.Println("部署文件已存在, 跳过")
			continue
		}
		fmt.Println("部署文件不存在, 创建中...")
		err = tools.File(deployPath).WriteString(getDeploymentXml(file.Name()))
		if err != nil {
			fmt.Println("写入部署xml失败: ", err.Error())
			continue
		}
		fmt.Println("写入部署xml成功")
	}
}

func getDeploymentXml(projectName string) string {
	pathNames := strings.Split(SftpName, ",")
	paths := ""
	for _, pathName := range pathNames {
		paths += getDeployPathNode(pathName, projectName)
	}

	xml := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<project version=\"4\">\n  <component name=\"PublishConfigData\" remoteFilesAllowedToDisappearOnAutoupload=\"false\">\n    <serverData>\n      {paths}    </serverData>\n  </component>\n</project>"
	xml = strings.ReplaceAll(xml, "{paths}", paths)
	return xml
}

func getDeployPathNode(sftpName string, projectName string) string {
	pathNode := "<paths name=\"{path name}\">\n        <serverdata>\n          <mappings>\n            <mapping deploy=\"{deploy}\" local=\"$PROJECT_DIR$\" web=\"/\" />\n          </mappings>\n        </serverdata>\n      </paths>\n"
	deploy := filepath.Join(ServerDir, projectName)
	pathNode = strings.Replace(pathNode, "{deploy}", deploy, -1)
	pathNode = strings.Replace(pathNode, "{path name}", sftpName, -1)
	return pathNode
}

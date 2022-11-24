package ezgin

type Config struct {
	App   App
	Nacos Nacos
	Gin   Gin
	Logs  Logs
	I18n  I18n
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
	Middleware string // gin中间件, 用逗号分隔, 暂时支持cors, trace, logs 不填则默认全部开启, - 表示不开启
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

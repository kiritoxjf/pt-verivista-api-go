package interfaces

// ErrorResponse 错误信息
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Config 配置文件
type Config struct {
	DB   DB
	Mail Mail
}

// DB 数据库
type DB struct {
	IP     string `json:"ip_addr"`
	Port   string `json:"port"`
	Driver string `json:"driver"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Name   string `json:"name"`
}

// Mail 邮件
type Mail struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	User string `json:"user"`
	Pwd  string `json:"pwd"`
	Nick string `json:"nick"`
}

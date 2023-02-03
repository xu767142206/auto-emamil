package mail

type Config struct {
	Addr     string `json:"addr"`     //地址 imap.qq.com
	Port     int    `json:"port"`     //端口 993
	Mail     string `json:"mail"`     //邮箱 /1368332201@qq.com
	Password string `json:"password"` //密码 beldkofrdjeuifdf
}

// NewConfig 初始化配置文件
func NewConfig() Config {
	return Config{
		Port: 993,
		Addr: "imap.qq.com",
	}
}

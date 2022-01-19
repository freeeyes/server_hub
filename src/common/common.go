package common

//IP信息
type Ip_info struct {
	Ip_   string
	Port_ string
}

type Net_io struct {
	Server_ip_   string
	Server_port_ string
}

//Json文件配置
type Server_json_info struct {
	Tcp_server_       Net_io
	Recv_queue_count_ int
}

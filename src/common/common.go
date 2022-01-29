package common

//IP信息
type Ip_info struct {
	Ip_   string
	Port_ string
}

//session接口
type Session_Info interface {
	Send_Io(data []byte, data_len int)
	Close_Io()
}

type Net_io struct {
	Server_ip_   string
	Server_port_ string
}

type Serial_io struct {
	Serial_session_id_ int
	Serial_name_       string
	Serial_frequency_  int
}

//Json文件配置
type Server_json_info struct {
	Tcp_server_       []Net_io
	Udp_Server_       []Net_io
	Serial_Server_    []Serial_io
	Recv_queue_count_ int
	Recv_buff_size_   int
	Send_buff_size_   int
}

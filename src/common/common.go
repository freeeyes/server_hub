package common

//IP信息
type Ip_info struct {
	Ip_   string
	Port_ string
}

//接续后的完整数据包数据
type Pakcet_info struct {
	Command_id_ uint16
	Data_       []byte
	Data_len_   uint32
}

//session接口
type Session_Info interface {
	Send_Io(data []byte, data_len int)
	Close_Io()
	Get_listen_Info() string
	Get_remote_info() string
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
	Io_time_check_    int
}

//关闭监听相关函数接口
type Io_Listen interface {
	Send_finish_listen_message()
}

//数据包事件关联接口
type Load_server_logic interface {
	Regedit_command(uint16, func(int, []byte, int, Session_Info)) bool
}

//逻辑模块接口
type Server_logic_info interface {
	Init(load_serevr_logic Load_server_logic)
}

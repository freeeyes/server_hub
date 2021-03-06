package server_logic

import (
	"fmt"
	"server_hub/common"
	"time"
)

type Io_Logic_info struct {
	Session_id_     int
	Server_ip_info_ string
	Client_ip_info_ string
	Session_info_   common.Session_Info
}

type Tcp_io_events_logic struct {
	Io_Logic_list_ map[int]*Io_Logic_info
}

func (tcp_io_events_logic *Tcp_io_events_logic) Init(load_serevr_logic common.Load_server_logic) {
	tcp_io_events_logic.Io_Logic_list_ = make(map[int]*Io_Logic_info)

	load_serevr_logic.Regedit_command(0, tcp_io_events_logic.Connect)
	load_serevr_logic.Regedit_command(1, tcp_io_events_logic.Disconnect)
	load_serevr_logic.Regedit_command(0x2101, tcp_io_events_logic.Recv_data_2101)
}

func (tcp_io_events_logic *Tcp_io_events_logic) Connect(session_id int, io_data []byte, data_len int, session_info common.Session_Info) {
	fmt.Println("[Tcp_io_events_logic::connect]session_id=", session_id, " is connect time:(", time.Now().String(), ") do")
	fmt.Println("[Tcp_io_events_logic::connect]listen=", session_info.Get_listen_Info(), ") do")
	fmt.Println("[Tcp_io_events_logic::connect]remote=", session_info.Get_remote_info(), ") do")

	//添加映射
	io_Logic_info := new(Io_Logic_info)
	io_Logic_info.Session_id_ = session_id
	io_Logic_info.Client_ip_info_ = session_info.Get_listen_Info()
	io_Logic_info.Server_ip_info_ = session_info.Get_remote_info()
	io_Logic_info.Session_info_ = session_info
	tcp_io_events_logic.Io_Logic_list_[session_id] = io_Logic_info
}

func (tcp_io_events_logic *Tcp_io_events_logic) Disconnect(session_id int, io_data []byte, data_len int, session_info common.Session_Info) {
	fmt.Println("[Tcp_io_events_logic::disconnect]session_id=", session_id, " is connect time:(", time.Now().String(), ") do")

	//删除映射
	delete(tcp_io_events_logic.Io_Logic_list_, session_id)
}

func (tcp_io_events_logic *Tcp_io_events_logic) Recv_data_2101(session_id int, io_data []byte, data_len int, session_info common.Session_Info) {
	fmt.Println("[Tcp_io_events_logic::Recv_data]session_id=", session_id, " recv(", data_len, ") time:(", time.Now().String(), ") do")

	//处理收到的数据
	if map_session_info, ok := tcp_io_events_logic.Io_Logic_list_[session_id]; ok {
		//如果数据存在，返回数据
		//packet_version := binary.LittleEndian.Uint16(io_data[0:2])
		//packet_command := binary.LittleEndian.Uint16(io_data[2:4])

		//fmt.Println("[do_chan_work]packet_version=", packet_version, "packet_command=", packet_command, ") do")
		map_session_info.Session_info_.Send_Io(io_data, data_len)

	} else {
		fmt.Println("[do_chan_work]session is no find=(", session_id, ")")
	}
}

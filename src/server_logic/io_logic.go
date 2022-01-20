package server_logic

import (
	"encoding/binary"
	"fmt"
	"server_hub/common"
	"time"
)

type Io_Logic_info struct {
	Session_id_     int
	Server_ip_info_ common.Ip_info
	Client_ip_info_ common.Ip_info
	Session_info_   common.Session_Info
}

type Tcp_io_events_logic struct {
	Io_Logic_list_ map[int]*Io_Logic_info
}

func (tcp_io_events_logic *Tcp_io_events_logic) Init() {
	tcp_io_events_logic.Io_Logic_list_ = make(map[int]*Io_Logic_info)
}

func (tcp_io_events_logic *Tcp_io_events_logic) Connect(session_id int, server_ip_info common.Ip_info, client_ip_info common.Ip_info, session_info common.Session_Info) bool {
	fmt.Println("[Tcp_io_events_logic::connect]session_id=", session_id, " is connect time:(", time.Now().String(), ") do")
	fmt.Println("[Tcp_io_events_logic::connect]server ip=", server_ip_info.Ip_, ":", server_ip_info.Port_, ") do")
	fmt.Println("[Tcp_io_events_logic::connect]server ip=", client_ip_info.Ip_, ":", client_ip_info.Port_, ") do")

	//添加映射
	io_Logic_info := new(Io_Logic_info)
	io_Logic_info.Session_id_ = session_id
	io_Logic_info.Client_ip_info_ = client_ip_info
	io_Logic_info.Server_ip_info_ = server_ip_info
	io_Logic_info.Session_info_ = session_info
	tcp_io_events_logic.Io_Logic_list_[session_id] = io_Logic_info
	return true
}

func (tcp_io_events_logic *Tcp_io_events_logic) Disconnect(session_id int) bool {
	fmt.Println("[Tcp_io_events_logic::disconnect]session_id=", session_id, " is connect time:(", time.Now().String(), ") do")

	//删除映射
	delete(tcp_io_events_logic.Io_Logic_list_, session_id)
	return true
}

func (tcp_io_events_logic *Tcp_io_events_logic) Recv_data(session_id int, io_data []byte, data_len int, session_info common.Session_Info) bool {
	fmt.Println("[Tcp_io_events_logic::Recv_data]session_id=", session_id, " recv(", data_len, ") time:(", time.Now().String(), ") do")

	//处理收到的数据
	if map_session_info, ok := tcp_io_events_logic.Io_Logic_list_[session_id]; ok {
		//如果数据存在，返回数据
		packet_version := binary.LittleEndian.Uint16(io_data[0:2])
		packet_command := binary.LittleEndian.Uint16(io_data[2:4])

		fmt.Println("[do_chan_work]packet_version=", packet_version, "packet_command=", packet_command, ") do")
		//session_info.Send_Io(io_data, data_len)
		map_session_info.Session_info_.Send_Io(io_data, data_len)

	} else {
		fmt.Println("[do_chan_work]session is no find=(", session_id, ")")
	}

	return true
}

package server_logic

import (
	"fmt"
	"server_hub/common"
	"time"
)

type Tcp_io_events_logic struct {
}

func (tcp_io_events_logic *Tcp_io_events_logic) Connect(session_id int, server_ip_info common.Ip_info, client_ip_info common.Ip_info) bool {
	fmt.Println("[Tcp_io_events_logic::connect]session_id=", session_id, " is connect time:(", time.Now().String(), ") do")
	fmt.Println("[Tcp_io_events_logic::connect]server ip=", server_ip_info.Ip_, ":", server_ip_info.Port_, ") do")
	fmt.Println("[Tcp_io_events_logic::connect]server ip=", client_ip_info.Ip_, ":", client_ip_info.Port_, ") do")
	return true
}

func (tcp_io_events_logic *Tcp_io_events_logic) Disconnect(session_id int) bool {
	fmt.Println("[Tcp_io_events_logic::disconnect]session_id=", session_id, " is connect time:(", time.Now().String(), ") do")
	return true
}

func (tcp_io_events_logic *Tcp_io_events_logic) Recv_data(session_id int, io_data []byte, data_len int) bool {
	fmt.Println("[Tcp_io_events_logic::Recv_data]session_id=", session_id, " recv(", data_len, ") time:(", time.Now().String(), ") do")
	return true
}

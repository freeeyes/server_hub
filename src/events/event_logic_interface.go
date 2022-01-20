package events

import "server_hub/common"

type Io_events interface {
	Init()
	Connect(session_id int, server_ip_info common.Ip_info, client_ip_info common.Ip_info, session_info common.Session_Info) bool
	Disconnect(session_id int) bool
	Recv_data(session_id int, io_data []byte, data_len int, session_info common.Session_Info) bool
}

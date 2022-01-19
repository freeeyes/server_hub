package events

import "server_hub/common"

type Io_events interface {
	Connect(session_id int, server_ip_info common.Ip_info, client_ip_info common.Ip_info) bool
	Disconnect(session_id int) bool
	Recv_data(session_id int, io_data []byte, data_len int) bool
}

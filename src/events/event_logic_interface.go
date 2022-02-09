package events

import "server_hub/common"

type Io_buff_to_packet interface {
	Recv_buff_to_packet(data []byte, data_len int) ([]*common.Pakcet_info, int, bool)
}

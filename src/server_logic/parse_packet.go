package server_logic

import (
	"encoding/binary"
	"server_hub/common"
)

//负责处理接收和发送数据包解析和加密
//add by freeeyes

type Io_buff_to_packet_logoc struct {
}

func (io_buff_to_packet_logoc *Io_buff_to_packet_logoc) Recv_buff_to_packet(data []byte, data_len int) ([]*common.Pakcet_info, int, bool) {
	var packet_list []*common.Pakcet_info

	var read_pos uint32 = 0

	for {
		//解析当前数据(如果数据包头不全，则直接返回)
		if uint32(data_len)-read_pos < 40 {
			break
		}
		//fmt.Println("[Recv_buff_to_packet]data_len-read_pos=", uint32(data_len)-read_pos)

		packet_size := binary.LittleEndian.Uint32(data[read_pos+4 : read_pos+8])

		//如果包长度大于最大缓冲长度，则返回失败
		if packet_size > 1024 {
			return packet_list, 0, false
		}

		//如果接收的数据长度不够，继续等待接收
		if packet_size > uint32(data_len)-read_pos {
			break
		}

		packet_info := new(common.Pakcet_info)
		packet_len := 40 + packet_size
		packet_info.Data_ = make([]byte, packet_len)
		end_pos := uint32(read_pos) + packet_len
		//fmt.Println("[Recv_buff_to_packet]read_pos=", read_pos, ",end_pos=", end_pos)
		copy(packet_info.Data_, data[read_pos:end_pos])

		//解析出对应的command_id
		packet_command := binary.LittleEndian.Uint16(packet_info.Data_[2:4])
		packet_info.Command_id_ = packet_command
		packet_info.Data_len_ = packet_len

		packet_list = append(packet_list, packet_info)
		read_pos += packet_len

	}

	return packet_list, int(read_pos), true
}

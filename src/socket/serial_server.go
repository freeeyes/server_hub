package socket

import (
	"fmt"
	"server_hub/events"
	"server_hub/serial"
)

type Serial_Server struct {
	serial_name_      string
	serial_frequency_ int
	chan_work_        *events.Chan_Work
	recv_buff_size_   int
	send_buff_size_   int
	packet_parse_     events.Io_buff_to_packet
	Session_          Serial_Session
}

func (serial_Server *Serial_Server) Listen(name string, frequency int, chan_work *events.Chan_Work, recv_buff_size int, send_buff_size int, packet_parse events.Io_buff_to_packet) uint16 {
	serial_Server.serial_name_ = name
	serial_Server.serial_frequency_ = frequency
	serial_Server.chan_work_ = chan_work
	serial_Server.recv_buff_size_ = recv_buff_size
	serial_Server.send_buff_size_ = send_buff_size

	c := &serial.Config{Name: serial_Server.serial_name_, Baud: serial_Server.serial_frequency_}
	s, err := serial.OpenPort(c)
	if err != nil {
		fmt.Println("[Serial_Server::listen]err=", err)
		return 1
	}

	serial_Server.packet_parse_ = packet_parse

	serial_Server.Session_.Init(1,
		serial_Server.serial_name_,
		serial_Server.serial_frequency_,
		serial_Server.recv_buff_size_,
		serial_Server.send_buff_size_,
		s)
	defer serial_Server.Session_.serial_port_.Close()

	for {
		recv_len, err := serial_Server.Session_.serial_port_.Read(serial_Server.Session_.Get_recv_buff())
		if err != nil {
			fmt.Println("[Serial_Server::listen]read data fail!", err)
		}

		fmt.Println("[Serial_Server::listen]session id=1 is recv, datalen=", recv_len)
		serial_Server.Session_.Set_write_len(recv_len)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := serial_Server.packet_parse_.Recv_buff_to_packet(serial_Server.Session_.Get_read_buff(), serial_Server.Session_.Get_write_buff())
		if !parse_is_ok {
			fmt.Println("[Serial_Server::listen]parse_is_ok=", parse_is_ok)
			break
		}

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = serial_Server.Session_.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet
			message.Message_Len_ = len(packet)
			message.Session_info_ = &serial_Server.Session_
			serial_Server.chan_work_.Add_Message(message)
		}

		serial_Server.Session_.Reset_read_buff(read_len)

	}

	//链接断开事件
	var message = new(events.Io_Info)
	message.Session_id_ = serial_Server.Session_.Get_Session_ID()
	message.Message_type_ = events.Io_Event_DisConnect
	serial_Server.chan_work_.Add_Message(message)

	return 0
}

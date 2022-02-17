package socket

import (
	"fmt"
	"server_hub/common"
	"server_hub/events"
	"server_hub/serial"
)

type Serial_server struct {
	session_counter_interface_ common.Session_counter_interface
	session_id_                int
	serial_name_               string
	serial_frequency_          int
	chan_work_                 *events.Chan_Work
	recv_buff_size_            int
	send_buff_size_            int
	packet_parse_              events.Io_buff_to_packet
	Session_                   Serial_Session
	listen_close_chan_         chan int
}

func (serial_Server *Serial_server) Send_finish_listen_message() {
	//发送消息，所有链接关闭结束。现在可以关闭监听了
	fmt.Println("[Serial_server::Send_finish_listen_message]send message listen close")
	serial_Server.listen_close_chan_ <- 1
}

func (serial_Server *Serial_server) Finial_Finish() {
	//监听关闭，回收相关资源

	//当全部客户端关闭执行完成后，执行消息通知，告知系统关闭
	serial_Server.listen_close_chan_ = make(chan int, 1)

	//发送监听结束消息
	var message = new(events.Io_Info)
	message.Session_id_ = 0
	message.Message_type_ = events.Io_Listen_Close
	message.Io_LIsten_Close_ = serial_Server
	serial_Server.chan_work_.Add_Message(message)

	for {
		data := <-serial_Server.listen_close_chan_
		if data == 1 {
			break
		}
	}

	fmt.Println("[Serial_server::Finial_Finish]close ok")
}

func (serial_Server *Serial_server) Listen(name string, frequency int, chan_work *events.Chan_Work, session_counter_interface common.Session_counter_interface, recv_buff_size int, send_buff_size int, packet_parse events.Io_buff_to_packet) uint16 {
	serial_Server.serial_name_ = name
	serial_Server.serial_frequency_ = frequency
	serial_Server.chan_work_ = chan_work
	serial_Server.recv_buff_size_ = recv_buff_size
	serial_Server.send_buff_size_ = send_buff_size
	serial_Server.session_counter_interface_ = session_counter_interface

	c := &serial.Config{Name: serial_Server.serial_name_, Baud: serial_Server.serial_frequency_}
	s, err := serial.OpenPort(c)
	if err != nil {
		fmt.Println("[Serial_server::listen]err=", err)
		return 1
	}

	serial_Server.packet_parse_ = packet_parse
	serial_Server.session_id_ = serial_Server.session_counter_interface_.Get_session_id()

	serial_Server.Session_.Init(serial_Server.session_id_,
		serial_Server.serial_name_,
		serial_Server.serial_frequency_,
		serial_Server.recv_buff_size_,
		serial_Server.send_buff_size_,
		s)
	//defer serial_Server.Session_.serial_port_.Close()

	for {
		recv_len, err := serial_Server.Session_.serial_port_.Read(serial_Server.Session_.Get_recv_buff())
		if err != nil {
			fmt.Println("[Serial_server::listen]read data fail!", err)
		}

		fmt.Println("[Serial_server::listen]session id=1 is recv, datalen=", recv_len)
		serial_Server.Session_.Set_write_len(recv_len)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := serial_Server.packet_parse_.Recv_buff_to_packet(serial_Server.Session_.Get_read_buff(), serial_Server.Session_.Get_write_buff())
		if !parse_is_ok {
			fmt.Println("[Serial_server::listen]parse_is_ok=", parse_is_ok)
			break
		}

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = serial_Server.Session_.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet.Data_
			message.Message_Len_ = packet.Data_len_
			message.Command_id_ = packet.Command_id_
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

func (serial_Server *Serial_server) Close() {
	//停止监听
	fmt.Println("[Serial_server::Close]server ip=", serial_Server.serial_name_)
	fmt.Println("[Serial_server::Close]server frequency=", serial_Server.serial_frequency_)

	serial_Server.Session_.serial_port_.Close()
}

type Serial_server_manager struct {
	serial_listen_list_        map[uint16]*Serial_server
	chan_work_                 *events.Chan_Work
	session_counter_interface_ common.Session_counter_interface
	recv_buff_size_            int
	send_buff_size_            int
	serial_server_count_       uint16
}

func (serial_server_manager *Serial_server_manager) Init(chan_work *events.Chan_Work, session_counter_interface common.Session_counter_interface, recv_buff_size int, send_buff_size int) {
	serial_server_manager.serial_listen_list_ = make(map[uint16]*Serial_server)
	serial_server_manager.chan_work_ = chan_work
	serial_server_manager.session_counter_interface_ = session_counter_interface
	serial_server_manager.recv_buff_size_ = recv_buff_size
	serial_server_manager.send_buff_size_ = send_buff_size
	serial_server_manager.serial_server_count_ = 1
}

func (serial_server_manager *Serial_server_manager) Listen(name string, frequency int, packet_parse events.Io_buff_to_packet) uint16 {
	curr_serial_server_count := serial_server_manager.serial_server_count_
	serial_server := new(Serial_server)
	serial_server_manager.serial_listen_list_[curr_serial_server_count] = serial_server
	serial_server_manager.serial_server_count_++

	go serial_server.Listen(name,
		frequency,
		serial_server_manager.chan_work_,
		serial_server_manager.session_counter_interface_,
		serial_server_manager.recv_buff_size_,
		serial_server_manager.send_buff_size_,
		packet_parse)

	return curr_serial_server_count
}

func (serial_server_manager *Serial_server_manager) Close() {
	for _, serial_server := range serial_server_manager.serial_listen_list_ {
		serial_server.Close()
	}

	fmt.Println("[Serial_server_manager::Close]close finish")
}

package socket

import (
	"fmt"
	"server_hub/serial"
	"strconv"
)

type Serial_Session struct {
	session_id_       int
	serial_name_      string
	serial_frequency_ int
	serial_port_      *serial.Port
	recv_buff_        []byte
	send_buff_        []byte
	recv_buff_size_   int
	send_buff_size_   int
	write_len_        int
}

func (serial_Session *Serial_Session) Init(session_id int, name string, frequency int, recv_buff_size int, send_buff_size int, serial_port *serial.Port) {
	serial_Session.session_id_ = session_id
	serial_Session.serial_name_ = name
	serial_Session.serial_frequency_ = frequency
	serial_Session.recv_buff_size_ = recv_buff_size
	serial_Session.send_buff_size_ = send_buff_size
	serial_Session.recv_buff_ = make([]byte, serial_Session.recv_buff_size_)
	serial_Session.send_buff_ = make([]byte, serial_Session.send_buff_size_)
	serial_Session.serial_port_ = serial_port
	serial_Session.write_len_ = 0
}

func (serial_Session *Serial_Session) Get_listen_Info() string {
	return serial_Session.serial_name_ + ":" + strconv.Itoa(serial_Session.serial_frequency_)
}

func (serial_Session *Serial_Session) Get_remote_info() string {
	return ""
}

func (serial_Session *Serial_Session) Get_write_buff() int {
	return serial_Session.write_len_
}

func (serial_Session *Serial_Session) Set_write_len(read_len int) {
	serial_Session.write_len_ += read_len
}

func (serial_Session *Serial_Session) Reset_read_buff(read_len int) {
	if read_len == serial_Session.write_len_ {
		//全部读完了
		serial_Session.write_len_ = 0
	} else {
		serial_Session.write_len_ -= read_len
	}
}

func (serial_Session *Serial_Session) Get_recv_buff() []byte {
	return serial_Session.recv_buff_[serial_Session.write_len_:(serial_Session.recv_buff_size_ - serial_Session.write_len_)]
}

func (serial_Session *Serial_Session) Get_read_buff() []byte {
	return serial_Session.recv_buff_
}

func (serial_Session *Serial_Session) Send_Io(data []byte, data_len int) bool {
	var send_len = 0

	for {
		if send_len == data_len {
			break
		}

		curr_send_len, err := serial_Session.serial_port_.Write(data[:data_len])
		if err != nil {
			//发送数据出错，打印错误信息
			fmt.Println("[Send_Io]session_id=", serial_Session.session_id_, "send err=", err)
			return false
		}
		send_len += curr_send_len
	}

	return true
}

func (serial_Session *Serial_Session) Close_Io() {
	serial_Session.serial_port_.Close()
}

func (serial_Session *Serial_Session) Get_Recv_Buff() []byte {
	return serial_Session.recv_buff_
}

func (serial_Session *Serial_Session) Get_Recv_Buff_size() int {
	return serial_Session.recv_buff_size_
}

func (serial_Session *Serial_Session) Get_Send_Buff() []byte {
	return serial_Session.send_buff_
}

func (serial_Session *Serial_Session) Get_Session_ID() int {
	return serial_Session.session_id_
}

func (serial_Session *Serial_Session) Show() {
	fmt.Println("[session]session id=", serial_Session.session_id_)
	fmt.Println("[session]serial name=", serial_Session.serial_name_)
	fmt.Println("[session]serial port=", serial_Session.serial_port_)
}

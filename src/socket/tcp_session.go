package socket

import (
	"fmt"
	"net"
)

type Tcp_Session struct {
	session_id_     int
	client_ip_      string
	client_port_    string
	server_ip_      string
	server_port_    string
	recv_buff_      []byte
	send_buff_      []byte
	recv_buff_size_ int
	recv_send_size_ int
	session_io_     net.Conn
}

func (tcp_session *Tcp_Session) Init(session_id int, client_ip string, client_port string, server_ip string, server_port string, session_io net.Conn) {
	tcp_session.session_id_ = session_id
	tcp_session.client_ip_ = client_ip
	tcp_session.client_port_ = client_port
	tcp_session.server_ip_ = server_ip
	tcp_session.server_port_ = server_port
	tcp_session.recv_buff_size_ = 256
	tcp_session.recv_send_size_ = 256
	tcp_session.recv_buff_ = make([]byte, tcp_session.recv_buff_size_)
	tcp_session.send_buff_ = make([]byte, tcp_session.recv_send_size_)
	tcp_session.session_io_ = session_io
}

func (tcp_session *Tcp_Session) Send_Io(data []byte, data_len int) {
	tcp_session.session_io_.Write(data[:data_len])
}

func (tcp_session *Tcp_Session) Get_Recv_Buff() []byte {
	return tcp_session.recv_buff_
}

func (tcp_session *Tcp_Session) Get_Recv_Buff_size() int {
	return tcp_session.recv_buff_size_
}

func (tcp_session *Tcp_Session) Get_Send_Buff() []byte {
	return tcp_session.send_buff_
}

func (tcp_session *Tcp_Session) Get_Session_ID() int {
	return tcp_session.session_id_
}

func (tcp_session *Tcp_Session) Show() {
	fmt.Println("[session]session id=", tcp_session.session_id_)
	fmt.Println("[session]client ip=", tcp_session.client_ip_)
	fmt.Println("[session]client port=", tcp_session.client_port_)
	fmt.Println("[session]server ip=", tcp_session.server_ip_)
	fmt.Println("[session]server port=", tcp_session.server_port_)
}

package socket

import (
	"net"
	"strconv"
)

type Udp_Session struct {
	session_id_     int
	client_ip_      string
	client_port_    int
	server_ip_      string
	server_port_    string
	send_buff_      []byte
	recv_buff_size_ int
	send_buff_size_ int
	socket_         *net.UDPConn
}

func (udp_session *Udp_Session) Get_Session_ID() int {
	return udp_session.session_id_
}

func (udp_session *Udp_Session) Init(session_id int, client_ip string, client_port int, server_ip string, server_port string, recv_buff_size int, send_buff_size int) {
	udp_session.session_id_ = session_id
	udp_session.client_ip_ = client_ip
	udp_session.client_port_ = client_port
	udp_session.server_ip_ = server_ip
	udp_session.server_port_ = server_port
	udp_session.recv_buff_size_ = recv_buff_size
	udp_session.send_buff_size_ = send_buff_size
	udp_session.send_buff_ = make([]byte, udp_session.send_buff_size_)
}

func (udp_session *Udp_Session) Close_Io() {
	//udp无法关闭
}

func (udp_session *Udp_Session) Get_listen_Info() string {
	return udp_session.server_ip_ + ":" + udp_session.server_port_
}

func (udp_session *Udp_Session) Get_remote_info() string {
	return udp_session.client_ip_ + ":" + strconv.Itoa(udp_session.client_port_)
}

func (udp_session *Udp_Session) Send_Io(data []byte, data_len int) {
	if data_len > 0 {
		udp_session.socket_.WriteToUDP(data, &net.UDPAddr{IP: net.ParseIP(udp_session.client_ip_), Port: udp_session.client_port_})
	}
}

package socket

import (
	"fmt"
	"net"
	"server_hub/common"
	"server_hub/events"
	"strconv"
	"strings"
)

type Udp_client struct {
	session_id_     int
	session_        *Tcp_Session
	client_ip_      string
	client_port_    string
	server_ip_      string
	server_port_    string
	client_conn_    *net.UDPConn
	chan_work_      *events.Chan_Work
	recv_buff_size_ int
	send_buff_size_ int
	packet_parse_   events.Io_buff_to_packet
	is_connect_     bool
}

func (udp_client *Udp_client) Close_Event() {
	//发送链接断开事件
	var message = new(events.Io_Info)
	message.Session_id_ = udp_client.session_.Get_Session_ID()
	message.Message_type_ = events.Io_Event_DisConnect
	udp_client.chan_work_.Add_Message(message)

	udp_client.client_conn_.Close() // 关闭TCP连接
}

func (udp_client *Udp_client) ReConnect() bool {
	if udp_client.is_connect_ {
		return false
	} else {
		//重连
		server_port, _ := strconv.Atoi(udp_client.server_port_)
		conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(udp_client.server_ip_), Port: server_port})
		if err != nil {
			fmt.Println("[udp_client::connct]connect error=", err)
			return false
		}

		udp_client.client_conn_ = conn
		udp_client.is_connect_ = true

		go udp_client.Read_buff()
		return true
	}
}

func (udp_client *Udp_client) Close() {
	udp_client.client_conn_.Close()
	udp_client.is_connect_ = false

	//发送链接断开消息
	udp_client.Close_Event()
}

func (udp_client *Udp_client) Finally_Close() {
	if udp_client.is_connect_ {
		udp_client.is_connect_ = false
	}
}

func (udp_client *Udp_client) Read_buff() {
	defer udp_client.Finally_Close()

	udp_client.client_ip_ = strings.Split(udp_client.client_conn_.LocalAddr().String(), ":")[0]
	udp_client.client_port_ = strings.Split(udp_client.client_conn_.LocalAddr().String(), ":")[1]

	udp_client.session_.Init(udp_client.session_id_,
		udp_client.client_ip_,
		udp_client.client_port_,
		udp_client.server_ip_,
		udp_client.server_port_,
		udp_client.client_conn_,
		udp_client.recv_buff_size_,
		udp_client.send_buff_size_)

	//添加链接建立事件
	var message = new(events.Io_Info)
	message.Session_id_ = udp_client.session_id_
	message.Message_type_ = events.Io_Event_Connect
	message.Session_info_ = udp_client.session_
	udp_client.chan_work_.Add_Message(message)

	for {
		reqLen, _, err := udp_client.client_conn_.ReadFromUDP(udp_client.session_.Get_recv_buff())
		if err != nil || reqLen == 0 {
			fmt.Println("[Udp_client::Handle_Connection]session id=", udp_client.session_.Get_Session_ID(), ",err=", err)
			break
		}

		fmt.Println("[Udp_client::Handle_Connection]session id=", udp_client.session_.Get_Session_ID(), " is recv, datalen=", reqLen)
		udp_client.session_.Set_write_len(reqLen)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := udp_client.packet_parse_.Recv_buff_to_packet(udp_client.session_.Get_read_buff(), udp_client.session_.Get_write_buff())
		//fmt.Println("[Handle_Connection]packet_list=", len(packet_list))
		//fmt.Println("[Handle_Connection]read_len=", read_len)
		if !parse_is_ok {
			fmt.Println("[Udp_client::Handle_Connection]parse_is_ok=", parse_is_ok)
			udp_client.Close()
		}

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = udp_client.session_.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet.Data_
			message.Message_Len_ = packet.Data_len_
			message.Command_id_ = packet.Command_id_
			message.Session_info_ = udp_client.session_
			udp_client.chan_work_.Add_Message(message)
		}

		udp_client.session_.Reset_read_buff(read_len)
		fmt.Println("[Udp_client::Handle_Connection]writelen=", udp_client.session_.Get_write_buff())
	}
}

func (udp_client *Udp_client) Get_session_id() int {
	return udp_client.session_id_
}

func (udp_client *Udp_client) Is_connect() bool {
	return udp_client.is_connect_
}

func (udp_client *Udp_client) Connct(session_id int, server_ip string, server_port int, recv_buff_size int, send_buff_size int, packet_parse events.Io_buff_to_packet, chan_work *events.Chan_Work) bool {
	udp_client.session_id_ = session_id
	udp_client.server_ip_ = server_ip
	udp_client.server_port_ = strconv.Itoa(server_port)
	udp_client.recv_buff_size_ = recv_buff_size
	udp_client.send_buff_size_ = send_buff_size
	udp_client.packet_parse_ = packet_parse
	udp_client.chan_work_ = chan_work
	udp_client.is_connect_ = false

	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP(udp_client.server_ip_), Port: server_port})
	if err != nil {
		fmt.Println("[udp_client::connct]connect error=", err)
		return false
	}

	udp_client.client_conn_ = conn
	udp_client.is_connect_ = true

	udp_client.session_ = new(Tcp_Session)
	go udp_client.Read_buff()

	return true
}

type Client_udp_manager struct {
	client_tcp_list_           map[int]*Udp_client
	chan_work_                 *events.Chan_Work
	session_counter_interface_ common.Session_counter_interface
	recv_buff_size_            int
	send_buff_size_            int
}

func (client_udp_manager *Client_udp_manager) Init(chan_work *events.Chan_Work, session_counter_interface common.Session_counter_interface, recv_buff_size int, send_buff_size int) {
	client_udp_manager.chan_work_ = chan_work
	client_udp_manager.session_counter_interface_ = session_counter_interface
	client_udp_manager.recv_buff_size_ = recv_buff_size
	client_udp_manager.send_buff_size_ = send_buff_size

	client_udp_manager.client_tcp_list_ = make(map[int]*Udp_client)
}

func (client_udp_manager *Client_udp_manager) Connect_tcp(server_ip string, server_port int, packet_parse events.Io_buff_to_packet) int {
	session_id := client_udp_manager.session_counter_interface_.Get_session_id()

	udp_client := new(Udp_client)

	udp_client.Connct(session_id,
		server_ip,
		server_port,
		client_udp_manager.recv_buff_size_,
		client_udp_manager.send_buff_size_,
		packet_parse,
		client_udp_manager.chan_work_)

	client_udp_manager.client_tcp_list_[session_id] = udp_client

	return session_id
}

func (client_udp_manager *Client_udp_manager) Time_Check() {
	for _, v := range client_udp_manager.client_tcp_list_ {
		if !v.Is_connect() {
			fmt.Println("[Client_udp_manager::Time_Check]session_id=", v.Get_session_id(), " is reconnec")
			v.ReConnect()
		}
	}
}

func (client_udp_manager *Client_udp_manager) Close_all() {
	for k, v := range client_udp_manager.client_tcp_list_ {
		if v.Is_connect() {
			fmt.Println("[Client_udp_manager::Close_all]session_id=", v.Get_session_id(), " is Close")
			v.Close()
		}

		delete(client_udp_manager.client_tcp_list_, k)
	}
}

func (client_udp_manager *Client_udp_manager) Close(session_id int) {
	if v, ok := client_udp_manager.client_tcp_list_[session_id]; ok {
		//找到了，关闭它
		v.Close()
		delete(client_udp_manager.client_tcp_list_, session_id)
	}
}

func (client_udp_manager *Client_udp_manager) Reconnect(session_id int) {
	if v, ok := client_udp_manager.client_tcp_list_[session_id]; ok {
		//找到了，重新链接它
		v.ReConnect()
	}
}

package socket

import (
	"fmt"
	"net"
	"server_hub/common"
	"server_hub/events"
	"strconv"
	"strings"
)

type Tcp_client struct {
	session_id_     int
	session_        *Tcp_Session
	client_ip_      string
	client_port_    string
	server_ip_      string
	server_port_    string
	client_conn_    net.Conn
	chan_work_      *events.Chan_Work
	recv_buff_size_ int
	send_buff_size_ int
	packet_parse_   common.Io_buff_to_packet
	is_connect_     bool
}

func (tcp_client *Tcp_client) Close_Event() {
	//发送链接断开事件
	var message = new(events.Io_Info)
	message.Session_id_ = tcp_client.session_.Get_Session_ID()
	message.Message_type_ = events.Io_Event_DisConnect
	tcp_client.chan_work_.Add_Message(message)

	tcp_client.client_conn_.Close() // 关闭TCP连接
}

func (tcp_client *Tcp_client) ReConnect() bool {
	if tcp_client.is_connect_ {
		return false
	} else {
		//重连
		conn, err := net.Dial("tcp", tcp_client.server_ip_+":"+tcp_client.server_port_)
		if err != nil {
			fmt.Println("[tcp_client::connct]connect error=", err)
			return false
		}

		tcp_client.client_conn_ = conn
		tcp_client.is_connect_ = true

		go tcp_client.Read_buff()
		return true
	}
}

func (tcp_client *Tcp_client) Close() {
	tcp_client.client_conn_.Close()
	tcp_client.is_connect_ = false

	//发送链接断开消息
	tcp_client.Close_Event()
}

func (tcp_client *Tcp_client) Finally_Close() {
	if tcp_client.is_connect_ {
		tcp_client.is_connect_ = false
	}
}

func (tcp_client *Tcp_client) Read_buff() {
	defer tcp_client.Finally_Close()

	tcp_client.client_ip_ = strings.Split(tcp_client.client_conn_.LocalAddr().String(), ":")[0]
	tcp_client.client_port_ = strings.Split(tcp_client.client_conn_.LocalAddr().String(), ":")[1]

	tcp_client.session_.Init(tcp_client.session_id_,
		tcp_client.client_ip_,
		tcp_client.client_port_,
		tcp_client.server_ip_,
		tcp_client.server_port_,
		tcp_client.client_conn_,
		tcp_client.recv_buff_size_,
		tcp_client.send_buff_size_)

	//添加链接建立事件
	var message = new(events.Io_Info)
	message.Session_id_ = tcp_client.session_id_
	message.Message_type_ = events.Io_Event_Connect
	message.Session_info_ = tcp_client.session_
	tcp_client.chan_work_.Add_Message(message)

	for {
		reqLen, err := tcp_client.client_conn_.Read(tcp_client.session_.Get_recv_buff())
		if err != nil || reqLen == 0 {
			fmt.Println("[Tcp_client::Handle_Connection]session id=", tcp_client.session_.Get_Session_ID(), ",err=", err)
			break
		}

		fmt.Println("[Tcp_client::Handle_Connection]session id=", tcp_client.session_.Get_Session_ID(), " is recv, datalen=", reqLen)
		tcp_client.session_.Set_write_len(reqLen)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := tcp_client.packet_parse_.Recv_buff_to_packet(tcp_client.session_.Get_read_buff(), tcp_client.session_.Get_write_buff())
		//fmt.Println("[Handle_Connection]packet_list=", len(packet_list))
		//fmt.Println("[Handle_Connection]read_len=", read_len)
		if !parse_is_ok {
			fmt.Println("[Tcp_client::Handle_Connection]parse_is_ok=", parse_is_ok)
			tcp_client.Close()
		}

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = tcp_client.session_.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet.Data_
			message.Message_Len_ = packet.Data_len_
			message.Command_id_ = packet.Command_id_
			message.Session_info_ = tcp_client.session_
			tcp_client.chan_work_.Add_Message(message)
		}

		tcp_client.session_.Reset_read_buff(read_len)
		fmt.Println("[Tcp_client::Handle_Connection]writelen=", tcp_client.session_.Get_write_buff())
	}
}

func (tcp_client *Tcp_client) Get_session_id() int {
	return tcp_client.session_id_
}

func (tcp_client *Tcp_client) Is_connect() bool {
	return tcp_client.is_connect_
}

func (tcp_client *Tcp_client) Connct(session_id int, server_ip string, server_port int, recv_buff_size int, send_buff_size int, packet_parse common.Io_buff_to_packet, chan_work *events.Chan_Work) bool {
	tcp_client.session_id_ = session_id
	tcp_client.server_ip_ = server_ip
	tcp_client.server_port_ = strconv.Itoa(server_port)
	tcp_client.recv_buff_size_ = recv_buff_size
	tcp_client.send_buff_size_ = send_buff_size
	tcp_client.packet_parse_ = packet_parse
	tcp_client.chan_work_ = chan_work
	tcp_client.is_connect_ = false

	conn, err := net.Dial("tcp", tcp_client.server_ip_+":"+tcp_client.server_port_)
	if err != nil {
		fmt.Println("[tcp_client::connct]connect error=", err)
		return false
	}

	tcp_client.client_conn_ = conn
	tcp_client.is_connect_ = true

	tcp_client.session_ = new(Tcp_Session)
	go tcp_client.Read_buff()

	return true
}

type Client_tcp_manager struct {
	client_tcp_list_           map[int]*Tcp_client
	chan_work_                 *events.Chan_Work
	session_counter_interface_ common.Session_counter_interface
	recv_buff_size_            int
	send_buff_size_            int
	client_close_chan_         chan int
}

func (client_tcp_manager *Client_tcp_manager) Init(chan_work *events.Chan_Work, session_counter_interface common.Session_counter_interface, recv_buff_size int, send_buff_size int) {
	client_tcp_manager.chan_work_ = chan_work
	client_tcp_manager.session_counter_interface_ = session_counter_interface
	client_tcp_manager.recv_buff_size_ = recv_buff_size
	client_tcp_manager.send_buff_size_ = send_buff_size

	client_tcp_manager.client_tcp_list_ = make(map[int]*Tcp_client)
}

func (client_tcp_manager *Client_tcp_manager) Connect_tcp(server_ip string, server_port int, packet_parse common.Io_buff_to_packet) int {
	session_id := client_tcp_manager.session_counter_interface_.Get_session_id()

	tcp_client := new(Tcp_client)

	tcp_client.Connct(session_id,
		server_ip,
		server_port,
		client_tcp_manager.recv_buff_size_,
		client_tcp_manager.send_buff_size_,
		packet_parse,
		client_tcp_manager.chan_work_)

	client_tcp_manager.client_tcp_list_[session_id] = tcp_client

	return session_id
}

func (client_tcp_manager *Client_tcp_manager) Time_Check() {
	for _, v := range client_tcp_manager.client_tcp_list_ {
		if !v.Is_connect() {
			fmt.Println("[Client_tcp_manager::Time_Check]session_id=", v.Get_session_id(), " is reconnec")
			v.ReConnect()
		}
	}
}

func (client_tcp_manager *Client_tcp_manager) Close_all() {
	for k, v := range client_tcp_manager.client_tcp_list_ {
		if v.Is_connect() {
			fmt.Println("[Client_tcp_manager::Close_all]session_id=", v.Get_session_id(), " is Close")
			v.Close()
		}

		delete(client_tcp_manager.client_tcp_list_, k)
	}

	//发送天庭消息结束信息
	client_tcp_manager.client_close_chan_ = make(chan int, 1)

	//发送监听结束消息
	var message = new(events.Io_Info)
	message.Session_id_ = 0
	message.Message_type_ = events.Io_Listen_Close
	message.Io_LIsten_Close_ = client_tcp_manager
	client_tcp_manager.chan_work_.Add_Message(message)

	for {
		data := <-client_tcp_manager.client_close_chan_
		if data == 1 {
			break
		}
	}

	fmt.Println("[Client_tcp_manager::Close_all] is finish")
}

func (client_tcp_manager *Client_tcp_manager) Close(session_id int) {
	if v, ok := client_tcp_manager.client_tcp_list_[session_id]; ok {
		//找到了，关闭它
		v.Close()
		delete(client_tcp_manager.client_tcp_list_, session_id)
	}
}

func (client_tcp_manager *Client_tcp_manager) Reconnect(session_id int) {
	if v, ok := client_tcp_manager.client_tcp_list_[session_id]; ok {
		//找到了，重连它
		v.ReConnect()
	}
}

func (client_tcp_manager *Client_tcp_manager) Send_finish_listen_message() {
	//发送消息，所有链接关闭结束。现在可以关闭监听了
	fmt.Println("[client_tcp_manager::Send_finish_listen_message]send message listen close")
	client_tcp_manager.client_close_chan_ <- 1
}

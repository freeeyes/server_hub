package socket

import (
	"fmt"
	"net"
	"server_hub/events"
	"server_hub/server_logic"
	"strings"
)

//实现tcp server监听的功能
//add by freeeyes

type Tcp_Serve struct {
	session_id_count_ int
	server_ip_        string
	server_port_      string
	listen_           net.Listener
	session_list_     map[int]*Tcp_Session
	chan_work_        *events.Chan_Work
	recv_buff_size_   int
	send_buff_size_   int
	packet_parse_     events.Io_buff_to_packet
}

func (tcp_server *Tcp_Serve) Handle_Connection(c net.Conn, packet_parse events.Io_buff_to_packet) {
	var session = new(Tcp_Session)
	session.Init(tcp_server.session_id_count_,
		strings.Split(c.RemoteAddr().String(), ":")[0],
		strings.Split(c.RemoteAddr().String(), ":")[1],
		tcp_server.server_ip_,
		tcp_server.server_port_,
		c,
		tcp_server.recv_buff_size_,
		tcp_server.send_buff_size_)

	//添加链接建立事件
	var message = new(events.Io_Info)
	message.Session_id_ = tcp_server.session_id_count_
	message.Message_type_ = events.Io_Event_Connect
	message.Server_ip_info_.Ip_ = session.server_ip_
	message.Server_ip_info_.Port_ = session.server_port_
	message.Client_ip_info_.Ip_ = session.client_ip_
	message.Client_ip_info_.Port_ = session.client_port_
	message.Session_info_ = session
	tcp_server.chan_work_.Add_Message(message)

	tcp_server.session_list_[session.Get_Session_ID()] = session

	tcp_server.session_id_count_++
	//session.Show()

	defer func() {
		//发送链接断开事件
		var message = new(events.Io_Info)
		message.Session_id_ = session.Get_Session_ID()
		message.Message_type_ = events.Io_Event_DisConnect
		tcp_server.chan_work_.Add_Message(message)

		c.Close()
		delete(tcp_server.session_list_, session.Get_Session_ID())
	}()

	for {
		reqLen, err := c.Read(session.Get_recv_buff())
		if err != nil || reqLen == 0 {
			fmt.Println(err)
			return
		}

		fmt.Println("[Handle_Connection]session id=", session.Get_Session_ID(), " is recv, datalen=", reqLen)
		session.Set_write_len(reqLen)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := packet_parse.Recv_buff_to_packet(session.Get_read_buff(), session.Get_write_buff())
		//fmt.Println("[Handle_Connection]packet_list=", len(packet_list))
		//fmt.Println("[Handle_Connection]read_len=", read_len)
		fmt.Println("[Handle_Connection]parse_is_ok=", parse_is_ok)

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = session.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet
			message.Message_Len_ = len(packet)
			message.Session_info_ = session
			tcp_server.chan_work_.Add_Message(message)
		}

		session.Reset_read_buff(read_len)
		fmt.Println("[Handle_Connection]writelen=", session.Get_write_buff())
	}
}

func (tcp_server *Tcp_Serve) Listen(ip string, port string, chan_work *events.Chan_Work, recv_buff_size int, send_buff_size int) uint16 {
	tcp_server.session_id_count_ = 1
	tcp_server.server_ip_ = ip
	tcp_server.server_port_ = port
	tcp_server.chan_work_ = chan_work
	tcp_server.recv_buff_size_ = recv_buff_size
	tcp_server.send_buff_size_ = send_buff_size

	//初始化解析接口
	tcp_server.packet_parse_ = new(server_logic.Io_buff_to_packet_logoc)

	//初始化map
	tcp_server.session_list_ = make(map[int]*Tcp_Session)

	//开始监听
	server_info := tcp_server.server_ip_ + ":" + tcp_server.server_port_
	l, err := net.Listen("tcp4", server_info)
	if err != nil {
		fmt.Println("[listen]err=", err)
		return 1
	}

	tcp_server.listen_ = l

	defer tcp_server.listen_.Close()

	tcp_server.Show()
	fmt.Println("[listen]success")
	for {
		c, err := tcp_server.listen_.Accept()
		if err != nil {
			fmt.Println("[accept]err=", err)
			break
		}

		//处理数据
		go tcp_server.Handle_Connection(c, tcp_server.packet_parse_)
	}

	return 0
}

func (tcp_server *Tcp_Serve) Show() {
	fmt.Println("[show]server ip=", tcp_server.server_ip_)
	fmt.Println("[show]server port=", tcp_server.server_port_)
}

func (tcp_server *Tcp_Serve) Close() {
	fmt.Println("[Close]server ip=", tcp_server.server_ip_)
	tcp_server.listen_.Close()
}

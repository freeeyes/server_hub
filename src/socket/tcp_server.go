package socket

import (
	"fmt"
	"net"
	"server_hub/events"
	"strings"
)

//实现tcp server监听的功能
//add by freeeyes

type Tcp_server struct {
	session_id_count_  int
	server_ip_         string
	server_port_       string
	listen_            net.Listener
	session_list_      map[int]*Tcp_Session
	chan_work_         *events.Chan_Work
	recv_buff_size_    int
	send_buff_size_    int
	packet_parse_      events.Io_buff_to_packet
	listen_close_chan_ chan int
}

func (tcp_server *Tcp_server) Handle_Connection(c net.Conn, packet_parse events.Io_buff_to_packet) {
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

		fmt.Println("[Tcp_Serve::Handle_Connection]session id=", session.Get_Session_ID(), " is recv, datalen=", reqLen)
		session.Set_write_len(reqLen)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := packet_parse.Recv_buff_to_packet(session.Get_read_buff(), session.Get_write_buff())
		//fmt.Println("[Handle_Connection]packet_list=", len(packet_list))
		//fmt.Println("[Handle_Connection]read_len=", read_len)
		if !parse_is_ok {
			fmt.Println("[Tcp_Serve::Handle_Connection]parse_is_ok=", parse_is_ok)
			break
		}

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = session.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet.Data_
			message.Message_Len_ = packet.Data_len_
			message.Command_id_ = packet.Command_id_
			message.Session_info_ = session
			tcp_server.chan_work_.Add_Message(message)
		}

		session.Reset_read_buff(read_len)
		fmt.Println("[Tcp_Serve::Handle_Connection]writelen=", session.Get_write_buff())
	}
}

func (tcp_server *Tcp_server) Send_finish_listen_message() {
	//发送消息，所有链接关闭结束。现在可以关闭监听了
	fmt.Println("[Tcp_Serve::Send_finish_listen_message]send message listen close")
	tcp_server.listen_close_chan_ <- 1
}

func (tcp_server *Tcp_server) Finial_Finish() {
	//监听关闭，回收相关资源

	//遍历map,并关闭链接客户端对象
	for k, v := range tcp_server.session_list_ {
		fmt.Println("[Tcp_Serve::Finial_Finish]close session id=", k)
		v.session_io_.Close()
	}

	//当全部客户端关闭执行完成后，执行消息通知，告知系统关闭
	tcp_server.listen_close_chan_ = make(chan int, 1)

	//发送监听结束消息
	var message = new(events.Io_Info)
	message.Session_id_ = 0
	message.Message_type_ = events.Io_Listen_Close
	message.Io_LIsten_Close_ = tcp_server
	tcp_server.chan_work_.Add_Message(message)

	for {
		data := <-tcp_server.listen_close_chan_
		if data == 1 {
			break
		}
	}

	fmt.Println("[Tcp_Serve::Finial_Finish]close ok")
}

func (tcp_server *Tcp_server) Listen(ip string, port string, chan_work *events.Chan_Work, recv_buff_size int, send_buff_size int, packet_parse events.Io_buff_to_packet) uint16 {
	tcp_server.session_id_count_ = 1
	tcp_server.server_ip_ = ip
	tcp_server.server_port_ = port
	tcp_server.chan_work_ = chan_work
	tcp_server.recv_buff_size_ = recv_buff_size
	tcp_server.send_buff_size_ = send_buff_size

	//初始化解析接口
	tcp_server.packet_parse_ = packet_parse

	//初始化map
	tcp_server.session_list_ = make(map[int]*Tcp_Session)

	//开始监听
	server_info := tcp_server.server_ip_ + ":" + tcp_server.server_port_
	l, err := net.Listen("tcp4", server_info)
	if err != nil {
		fmt.Println("[Tcp_Serve::listen]err=", err)
		return 1
	}

	tcp_server.listen_ = l

	defer tcp_server.Finial_Finish()

	tcp_server.Show()
	fmt.Println("[Tcp_Serve::listen]success")
	for {
		c, err := tcp_server.listen_.Accept()
		if err != nil {
			fmt.Println("[Tcp_Serve::accept]err=", err)
			break
		}

		//处理数据
		go tcp_server.Handle_Connection(c, tcp_server.packet_parse_)
	}

	return 0
}

func (tcp_server *Tcp_server) Show() {
	fmt.Println("[Tcp_Serve::show]server ip=", tcp_server.server_ip_)
	fmt.Println("[Tcp_Serve::show]server port=", tcp_server.server_port_)
}

func (tcp_server *Tcp_server) Close() {
	fmt.Println("[Tcp_Serve::Close]server ip=", tcp_server.server_ip_)
	fmt.Println("[Tcp_Serve::Close]server port=", tcp_server.server_port_)
	tcp_server.listen_.Close()
}

package socket

import (
	"fmt"
	"net"
	"server_hub/events"
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
}

func (tcp_server *Tcp_Serve) Handle_Connection(c net.Conn) {
	var session = new(Tcp_Session)
	session.Init(tcp_server.session_id_count_,
		strings.Split(c.RemoteAddr().String(), ":")[0],
		strings.Split(c.RemoteAddr().String(), ":")[1],
		tcp_server.server_ip_,
		tcp_server.server_port_,
		c)

	//添加链接建立事件
	var message = new(events.Io_Info)
	message.Session_id_ = tcp_server.session_id_count_
	message.Message_type_ = events.Io_Event_Connect
	message.Server_ip_info_.Ip_ = session.server_ip_
	message.Server_ip_info_.Port_ = session.server_port_
	message.Client_ip_info_.Ip_ = session.client_ip_
	message.Client_ip_info_.Port_ = session.client_port_
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
		reqLen, err := c.Read(session.Get_Recv_Buff())
		if err != nil || reqLen == 0 {
			fmt.Println(err)
			return
		}

		//fmt.Println("[Handle_Connection]session id=", session.Get_Session_ID(), " is recv, datalen=", reqLen)

		//添加数据到达的消息
		var read_buffer = make([]byte, reqLen)
		copy(read_buffer, session.Get_Recv_Buff()[0:reqLen])

		var message = new(events.Io_Info)
		message.Session_id_ = tcp_server.session_id_count_
		message.Message_type_ = events.Io_Event_Data
		message.Mesaage_data_ = read_buffer
		message.Message_Len_ = reqLen
		message.Session_info_ = session
		tcp_server.chan_work_.Add_Message(message)
	}
}

func (tcp_server *Tcp_Serve) Listen(ip string, port string, chan_work *events.Chan_Work) uint16 {
	tcp_server.session_id_count_ = 1
	tcp_server.server_ip_ = ip
	tcp_server.server_port_ = port
	tcp_server.chan_work_ = chan_work

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
		go tcp_server.Handle_Connection(c)
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

package socket

import (
	"fmt"
	"net"
	"server_hub/events"
	"strconv"
)

type Udp_Serve struct {
	session_id_count_  int
	server_ip_         string
	server_port_       string
	listen_            *net.UDPConn
	session_list_      map[string]*Udp_Session
	chan_work_         *events.Chan_Work
	recv_buff_size_    int
	send_buff_size_    int
	packet_parse_      events.Io_buff_to_packet
	listen_close_chan_ chan int
}

func (udp_server *Udp_Serve) Send_finish_listen_message() {
	//发送消息，所有链接关闭结束。现在可以关闭监听了
	fmt.Println("[Udp_Serve::Send_finish_listen_message]send message listen close")
	udp_server.listen_close_chan_ <- 1
}

func (udp_server *Udp_Serve) Finial_Finish() {
	//监听关闭，回收相关资源

	//遍历map,并关闭链接客户端对象
	for k, v := range udp_server.session_list_ {
		fmt.Println("[Udp_Serve::Finial_Finish]close session id=", k)
		var message = new(events.Io_Info)
		message.Session_id_ = v.Get_Session_ID()
		message.Message_type_ = events.Io_Event_DisConnect
		udp_server.chan_work_.Add_Message(message)
	}

	//当全部客户端关闭执行完成后，执行消息通知，告知系统关闭
	udp_server.listen_close_chan_ = make(chan int, 1)

	//发送监听结束消息
	var message = new(events.Io_Info)
	message.Session_id_ = 0
	message.Message_type_ = events.Io_Listen_Close
	message.Io_LIsten_Close_ = udp_server
	udp_server.chan_work_.Add_Message(message)

	for {
		data := <-udp_server.listen_close_chan_
		if data == 1 {
			break
		}
	}

	fmt.Println("[Udp_Serve::Finial_Finish]close ok")
}

func (udp_server *Udp_Serve) Show() {
	fmt.Println("[Udp_Serve::show]server ip=", udp_server.server_ip_)
	fmt.Println("[Udp_Serve::show]server port=", udp_server.server_port_)
}

func (udp_server *Udp_Serve) Recv_udp_data() uint16 {
	recv_data := make([]byte, udp_server.recv_buff_size_)
	for {
		// 读取数据
		recv_len, client_addr, err := udp_server.listen_.ReadFromUDP(recv_data)
		if err != nil {
			fmt.Println("[Udp_Serve::listen]read data fail!", err)
			break
		}

		//在map里寻找
		client_ip := client_addr.IP.String()
		client_port := client_addr.Port
		client_key := client_ip + ":" + strconv.Itoa(client_port)

		session, ok := udp_server.session_list_[client_key]
		if !ok {
			//不存在
			session = new(Udp_Session)
			session.Init(udp_server.session_id_count_,
				client_ip,
				client_port,
				udp_server.server_ip_,
				udp_server.server_port_,
				udp_server.recv_buff_size_,
				udp_server.send_buff_size_)
			udp_server.session_id_count_++
			udp_server.session_list_[client_key] = session

			//添加链接建立事件
			var message = new(events.Io_Info)
			message.Session_id_ = session.session_id_
			message.Message_type_ = events.Io_Event_Connect
			message.Server_ip_info_.Ip_ = session.server_ip_
			message.Server_ip_info_.Port_ = session.server_port_
			message.Client_ip_info_.Ip_ = session.client_ip_
			message.Client_ip_info_.Port_ = strconv.Itoa(session.client_port_)
			udp_server.chan_work_.Add_Message(message)
		}

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := udp_server.packet_parse_.Recv_buff_to_packet(recv_data, recv_len)

		if read_len != recv_len || !parse_is_ok {
			fmt.Println("[Udp_Serve::listen]packet parse data fail!")

			//发送链接断开事件
			var message = new(events.Io_Info)
			message.Session_id_ = session.Get_Session_ID()
			message.Message_type_ = events.Io_Event_DisConnect
			udp_server.chan_work_.Add_Message(message)
		} else {
			for _, packet := range packet_list {
				var message = new(events.Io_Info)
				message.Session_id_ = session.Get_Session_ID()
				message.Message_type_ = events.Io_Event_Data
				message.Mesaage_data_ = packet
				message.Message_Len_ = len(packet)
				message.Session_info_ = session
				udp_server.chan_work_.Add_Message(message)
			}
		}
	}

	return 0
}

func (udp_server *Udp_Serve) Listen(ip string, port string, chan_work *events.Chan_Work, recv_buff_size int, send_buff_size int, packet_parse events.Io_buff_to_packet) uint16 {
	udp_server.session_id_count_ = 1
	udp_server.server_ip_ = ip
	udp_server.server_port_ = port
	udp_server.chan_work_ = chan_work
	udp_server.recv_buff_size_ = recv_buff_size
	udp_server.send_buff_size_ = send_buff_size

	//初始化解析接口
	udp_server.packet_parse_ = packet_parse

	//初始化map
	udp_server.session_list_ = make(map[string]*Udp_Session)

	//开始监听
	server_port, _ := strconv.Atoi(udp_server.server_port_)
	l, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.ParseIP(udp_server.server_ip_), Port: server_port})
	if err != nil {
		fmt.Println("[Udp_Serve::listen]err=", err)
		return 1
	}

	udp_server.listen_ = l

	defer udp_server.Finial_Finish()

	udp_server.Show()
	fmt.Println("[Udp_Serve::listen]success")

	//开始接收数据
	return udp_server.Recv_udp_data()
}

func (udp_server *Udp_Serve) Close() {
	fmt.Println("[Udp_Serve::Close]server ip=", udp_server.server_ip_)
	fmt.Println("[Udp_Serve::Close]server port=", udp_server.server_ip_)
	udp_server.listen_.Close()
}

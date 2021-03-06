package socket

import (
	"fmt"
	"net"
	"server_hub/common"
	"server_hub/events"
	"strconv"
)

type Udp_server struct {
	Session_counter_interface_ common.Session_counter_interface
	server_ip_                 string
	server_port_               string
	listen_                    *net.UDPConn
	session_list_              map[string]*Udp_Session
	chan_work_                 *events.Chan_Work
	recv_buff_size_            int
	send_buff_size_            int
	packet_parse_              common.Io_buff_to_packet
	listen_close_chan_         chan int
}

func (udp_server *Udp_server) Send_finish_listen_message() {
	//发送消息，所有链接关闭结束。现在可以关闭监听了
	fmt.Println("[Udp_Serve::Send_finish_listen_message]send message listen close")
	udp_server.listen_close_chan_ <- 1
}

func (udp_server *Udp_server) Finial_Finish() {
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

func (udp_server *Udp_server) Show() {
	fmt.Println("[Udp_Serve::show]server ip=", udp_server.server_ip_)
	fmt.Println("[Udp_Serve::show]server port=", udp_server.server_port_)
}

func (udp_server *Udp_server) Recv_udp_data() uint16 {
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
			session_id := udp_server.Session_counter_interface_.Get_session_id()

			//不存在
			session = new(Udp_Session)
			session.Init(session_id,
				client_ip,
				client_port,
				udp_server.server_ip_,
				udp_server.server_port_,
				udp_server.recv_buff_size_,
				udp_server.send_buff_size_)
			udp_server.session_list_[client_key] = session

			//添加链接建立事件
			var message = new(events.Io_Info)
			message.Session_id_ = session.session_id_
			message.Message_type_ = events.Io_Event_Connect
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
				message.Mesaage_data_ = packet.Data_
				message.Message_Len_ = packet.Data_len_
				message.Command_id_ = packet.Command_id_
				message.Session_info_ = session
				udp_server.chan_work_.Add_Message(message)
			}
		}
	}

	return 0
}

func (udp_server *Udp_server) Listen(ip string, port string, chan_work *events.Chan_Work, session_counter_interface common.Session_counter_interface, recv_buff_size int, send_buff_size int, packet_parse common.Io_buff_to_packet) uint16 {
	udp_server.server_ip_ = ip
	udp_server.server_port_ = port
	udp_server.chan_work_ = chan_work
	udp_server.recv_buff_size_ = recv_buff_size
	udp_server.send_buff_size_ = send_buff_size
	udp_server.Session_counter_interface_ = session_counter_interface

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

func (udp_server *Udp_server) Close() {
	fmt.Println("[Udp_Serve::Close]server ip=", udp_server.server_ip_)
	fmt.Println("[Udp_Serve::Close]server port=", udp_server.server_ip_)
	udp_server.listen_.Close()
}

type Udp_server_manager struct {
	udp_listen_list_           map[uint16]*Udp_server
	chan_work_                 *events.Chan_Work
	session_counter_interface_ common.Session_counter_interface
	recv_buff_size_            int
	send_buff_size_            int
	udp_server_count_          uint16
}

func (udp_server_manager *Udp_server_manager) Init(chan_work *events.Chan_Work, session_counter_interface common.Session_counter_interface, recv_buff_size int, send_buff_size int) {
	udp_server_manager.udp_listen_list_ = make(map[uint16]*Udp_server)
	udp_server_manager.chan_work_ = chan_work
	udp_server_manager.session_counter_interface_ = session_counter_interface
	udp_server_manager.recv_buff_size_ = recv_buff_size
	udp_server_manager.send_buff_size_ = send_buff_size
	udp_server_manager.udp_server_count_ = 1
}

func (udp_server_manager *Udp_server_manager) Listen(ip string, port string, packet_parse common.Io_buff_to_packet) uint16 {
	curr_udp_server_count := udp_server_manager.udp_server_count_
	udp_server := new(Udp_server)
	udp_server_manager.udp_listen_list_[curr_udp_server_count] = udp_server
	udp_server_manager.udp_server_count_++

	go udp_server.Listen(ip,
		port,
		udp_server_manager.chan_work_,
		udp_server_manager.session_counter_interface_,
		udp_server_manager.recv_buff_size_,
		udp_server_manager.send_buff_size_,
		packet_parse)

	return curr_udp_server_count
}

func (udp_server_manager *Udp_server_manager) Close() {
	for _, udp_server := range udp_server_manager.udp_listen_list_ {
		udp_server.Close()
	}

	fmt.Println("[Udp_server_manager::Close]close finish")
}

package main

//测试接口代码编写
//add by freeeyes

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"server_hub/common"
	"server_hub/events"
	"server_hub/server_logic"
	"server_hub/socket"
	"syscall"
	"time"
)

//监控信号量
func Catch_sig(ch chan os.Signal, chan_work *events.Chan_Work, io_manager *events.Io_manager) {
	sig := <-ch
	fmt.Println("\nsig received:", sig)

	switch sig {
	case syscall.SIGINT:
		fmt.Println("handling a SIGINT now!")
	case syscall.SIGTERM:
		fmt.Println("handling a SIGTERM in an entirely different way!")
	default:
		fmt.Println("unexpected signal received")
	}

	defer Close(io_manager, chan_work)

	// 终止(发送chan消息)
	//var message = new(events.Io_Info)
	//message.Message_type_ = events.Io_Exit
	//chan_work.Add_Message(message)
}

func Close(io_manager *events.Io_manager, chan_work *events.Chan_Work) {
	io_manager.Listen_tcp_manager_.Close()
	io_manager.Listen_udp_manager_.Close()
	io_manager.Listen_serial_manager_.Close()
	io_manager.Client_tcp_manager_.Close_all()
	io_manager.Client_udp_manager_.Close_all()

	//发送结束chan消息
	var message = new(events.Io_Info)
	message.Message_type_ = events.Io_Finish
	chan_work.Add_Message(message)
}

//读取配置文件
func Read_server_json(config_file_path string, server_json_info interface{}) bool {
	data, err := ioutil.ReadFile(config_file_path)
	if err != nil {
		fmt.Println("[read_server_json]no find file")
		return false
	}

	err = json.Unmarshal(data, &server_json_info)
	if err != nil {
		fmt.Println("[read_server_json]json is error(", err, ")")
		return false
	}

	return true
}

func Show_config(server_json_info common.Server_json_info) {
	for _, tcp_server_config := range server_json_info.Tcp_server_ {
		fmt.Println("[read_server_json]tcp_ip=", tcp_server_config.Server_ip_)
		fmt.Println("[read_server_json]tcp_port=", tcp_server_config.Server_port_)
		fmt.Println("[read_server_json]===================")
	}

	for _, tcp_server_config := range server_json_info.Udp_Server_ {
		fmt.Println("[read_server_json]udp_ip=", tcp_server_config.Server_ip_)
		fmt.Println("[read_server_json]udp_port=", tcp_server_config.Server_port_)
		fmt.Println("[read_server_json]===================")
	}

	fmt.Println("[read_server_json]Recv_buff_size_=", server_json_info.Recv_buff_size_)
	fmt.Println("[read_server_json]Send_buff_size_=", server_json_info.Send_buff_size_)
}

func main() {
	if len(os.Args) > 2 {
		fmt.Println("[main]args error.")
		return
	}

	config_file := "../config/server_config.json"
	if len(os.Args) == 2 {
		//记录对应的路径名
		config_file = os.Args[1]
	}

	io_manager := new(events.Io_manager)

	//读取配置文件
	server_json_info := common.Server_json_info{}
	if !Read_server_json(config_file, &server_json_info) {
		return
	}

	//启动IO事件处理线程
	chan_work := new(events.Chan_Work)

	//启动计数器
	var session_counter_interface common.Session_counter_interface = new(common.Session_counter_manager)
	session_counter_interface.Init()

	//创建对应的监听队列
	tcp_listen_list := new(socket.Tcp_server_manager)
	udp_listen_list := new(socket.Udp_server_manager)
	serial_listen_list := new(socket.Serial_server_manager)
	tcp_listen_list.Init(chan_work, session_counter_interface, server_json_info.Recv_buff_size_, server_json_info.Send_buff_size_)
	udp_listen_list.Init(chan_work, session_counter_interface, server_json_info.Recv_buff_size_, server_json_info.Send_buff_size_)
	serial_listen_list.Init(chan_work, session_counter_interface, server_json_info.Recv_buff_size_, server_json_info.Send_buff_size_)
	io_manager.Set_tcp_manager(tcp_listen_list)
	io_manager.Set_udp_manager(udp_listen_list)
	io_manager.Set_serial_manager(serial_listen_list)

	//显示配置文件内容
	Show_config(server_json_info)

	// 初始化通道
	signals := make(chan os.Signal, 1)

	// 将它们连接到信号lib
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	//初始化收发队列
	chan_work.Start(server_json_info.Recv_queue_count_, server_json_info.Io_time_check_)

	//创建消息解析类
	var packet_parse events.Io_buff_to_packet = new(server_logic.Io_buff_to_packet_logoc)

	//启动tcp监听
	for _, tcp_server_config := range server_json_info.Tcp_server_ {
		tcp_listen_list.Listen(tcp_server_config.Server_ip_,
			tcp_server_config.Server_port_,
			packet_parse)
	}

	//启动udp监听
	for _, tcp_server_config := range server_json_info.Udp_Server_ {
		udp_listen_list.Listen(tcp_server_config.Server_ip_,
			tcp_server_config.Server_port_,
			packet_parse)
	}

	//启动serial监听
	for _, serial_server_config := range server_json_info.Serial_Server_ {
		serial_listen_list.Listen(serial_server_config.Serial_name_,
			serial_server_config.Serial_frequency_,
			packet_parse)
	}

	//启动tcp服务器间链接
	var client_tcp_manager = new(socket.Client_tcp_manager)
	client_tcp_manager.Init(chan_work,
		session_counter_interface,
		server_json_info.Recv_buff_size_,
		server_json_info.Send_buff_size_)
	io_manager.Set_client_tcp_manager(client_tcp_manager)

	//驱动udp服务器间链接
	var client_udp_manager = new(socket.Client_udp_manager)
	client_udp_manager.Init(chan_work,
		session_counter_interface,
		server_json_info.Recv_buff_size_,
		server_json_info.Send_buff_size_)
	io_manager.Set_client_udp_manager(client_udp_manager)

	chan_work.Add_tcp_client_manager(client_tcp_manager)
	chan_work.Add_udp_client_manager(client_udp_manager)
	chan_work.Add_listen_tcp_manager(tcp_listen_list)
	chan_work.Add_listen_udp_manager(udp_listen_list)
	chan_work.Add_listen_serial_manager(serial_listen_list)

	//启动定时器
	if server_json_info.Io_time_check_ > 0 {
		timeTickerChan := time.Tick(time.Millisecond * time.Duration(server_json_info.Io_time_check_))
		go func() {
			for {
				//发送自检消息
				var message = new(events.Io_Info)
				message.Message_type_ = events.Timer_Check
				chan_work.Add_Message(message)
				<-timeTickerChan
			}
		}()
	}

	//测试关闭监听流程
	//time.Sleep(1 * time.Second)
	//listen_group.tcp_listen_list_[0].Close()
	go Catch_sig(signals, chan_work, io_manager)

	//开始运行
	chan_work.Run()
}

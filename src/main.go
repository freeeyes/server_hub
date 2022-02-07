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
)

type Listen_group struct {
	tcp_listen_list_    map[int]*socket.Tcp_server
	udp_listen_list_    map[int]*socket.Udp_Serve
	serial_listen_list_ map[int]*socket.Serial_Server
}

func (listen_group *Listen_group) Init() {
	listen_group.tcp_listen_list_ = make(map[int]*socket.Tcp_server)
	listen_group.udp_listen_list_ = make(map[int]*socket.Udp_Serve)
	listen_group.serial_listen_list_ = make(map[int]*socket.Serial_Server)
}

//监控信号量
func Catch_sig(ch chan os.Signal, done chan bool) {
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

	// 终止
	done <- true
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

func Close(listen_group *Listen_group) {
	//关闭服务器，释放资源
	for _, tcp_server := range listen_group.tcp_listen_list_ {
		tcp_server.Close()
	}

	for _, udp_server := range listen_group.udp_listen_list_ {
		udp_server.Close()
	}

	for _, serial_server := range listen_group.serial_listen_list_ {
		serial_server.Close()
	}
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
	//读取配置文件
	server_json_info := common.Server_json_info{}
	if !Read_server_json("../config/server_config.json", &server_json_info) {
		return
	}

	//创建对应的队列
	var listen_group = new(Listen_group)
	listen_group.Init()

	//显示配置文件内容
	Show_config(server_json_info)

	// 初始化通道
	signals := make(chan os.Signal, 1)
	done := make(chan bool)

	// 将它们连接到信号lib
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go Catch_sig(signals, done)

	//启动IO事件处理线程
	chan_work_ := new(events.Chan_Work)

	//初始化收发队列
	chan_work_.Start(server_json_info.Recv_queue_count_)

	//创建消息解析类
	var packet_parse events.Io_buff_to_packet = new(server_logic.Io_buff_to_packet_logoc)

	//启动tcp监听
	listen_id := 0
	for _, tcp_server_config := range server_json_info.Tcp_server_ {
		var tcp_server = new(socket.Tcp_server)

		listen_group.tcp_listen_list_[listen_id] = tcp_server
		listen_id++

		go tcp_server.Listen(tcp_server_config.Server_ip_,
			tcp_server_config.Server_port_,
			chan_work_,
			server_json_info.Recv_buff_size_,
			server_json_info.Send_buff_size_,
			packet_parse)

	}

	//启动udp监听
	for _, tcp_server_config := range server_json_info.Udp_Server_ {
		var udp_server = new(socket.Udp_Serve)

		listen_group.udp_listen_list_[listen_id] = udp_server
		listen_id++

		go udp_server.Listen(tcp_server_config.Server_ip_,
			tcp_server_config.Server_port_,
			chan_work_,
			server_json_info.Recv_buff_size_,
			server_json_info.Send_buff_size_,
			packet_parse)
	}

	//启动serial监听
	for _, serial_server_config := range server_json_info.Serial_Server_ {
		var serial_Server = new(socket.Serial_Server)

		listen_group.serial_listen_list_[listen_id] = serial_Server
		listen_id++

		go serial_Server.Listen(serial_server_config.Serial_session_id_,
			serial_server_config.Serial_name_,
			serial_server_config.Serial_frequency_,
			chan_work_,
			server_json_info.Recv_buff_size_,
			server_json_info.Send_buff_size_,
			packet_parse)
	}

	//测试关闭监听流程
	//time.Sleep(1 * time.Second)
	//listen_group.tcp_listen_list_[0].Close()

	defer Close(listen_group)

	<-done
	fmt.Println("Done!")
}

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
	"server_hub/socket"
	"syscall"
)

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

func main() {
	//读取配置文件
	server_json_info := common.Server_json_info{}
	if !Read_server_json("../config/server_config.json", &server_json_info) {
		return
	}

	fmt.Println("[read_server_json]tcp_ip=", server_json_info.Tcp_server_.Server_ip_)
	fmt.Println("[read_server_json]tcp_port=", server_json_info.Tcp_server_.Server_port_)
	fmt.Println("[read_server_json]recv_chan_count=", server_json_info.Recv_queue_count_)

	// 初始化通道
	signals := make(chan os.Signal, 1)
	done := make(chan bool)

	// 将它们连接到信号lib
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go Catch_sig(signals, done)

	//启动tcp监听
	var tcp_server = new(socket.Tcp_Serve)

	go tcp_server.Listen(server_json_info.Tcp_server_.Server_ip_,
		server_json_info.Tcp_server_.Server_port_,
		server_json_info.Recv_queue_count_)

	<-done
	fmt.Println("Done!")
}

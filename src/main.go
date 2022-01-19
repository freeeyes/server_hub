package main

//测试接口代码编写
//add by freeeyes

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"server_hub/events"
	"server_hub/socket"
	"syscall"
	"time"
)

//测试管道
func f_pipe_test() {
	reader, writter, err := os.Pipe()
	if err != nil {
		fmt.Printf("[f_pipe_test]pipe error: %s.\n", err)
	}

	go func() {
		time.Sleep(time.Second)
		writter.Write([]byte(time.Now().String()))
		time.Sleep(time.Second)
		writter.Close()
	}()

	for {
		datarread := make([]byte, 256)
		n, err := reader.Read(datarread)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("[f_pipe_test]read pipe error: %s.\n", err)
				break
			} else {
				fmt.Printf("[f_pipe_test]write is close.\n")
				break
			}
		}

		fmt.Printf("[f_pipe_test]time: %s.\n", string(datarread[:n]))
	}
}

//测试信号量
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

func main() {
	f_pipe_test()

	var man events.Persion = new(events.Man)
	man.Say()
	man.Eat("bread")

	var woman events.Persion = new(events.WoMan)
	woman.Say()
	woman.Eat("bread")

	// 初始化通道
	signals := make(chan os.Signal, 1)
	done := make(chan bool)

	// 将它们连接到信号lib
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go Catch_sig(signals, done)

	//启动tcp监听
	var tcp_server = new(socket.Tcp_Serve)

	go tcp_server.Listen("127.0.0.1", "10020", 100)

	<-done
	fmt.Println("Done!")
}

package events

import (
	"fmt"
	"server_hub/common"
	"server_hub/server_logic"
	"sync"
	"time"
)

//消息宏
const (
	Io_Event_Connect int = iota
	Io_Event_DisConnect
	Io_Event_Data
	Io_Exit
)

//数据消息包
type Io_Info struct {
	Session_id_     int
	Server_ip_info_ common.Ip_info
	Client_ip_info_ common.Ip_info
	Message_type_   int
	Mesaage_data_   []byte
	Message_Len_    int
	Session_info_   common.Session_Info
}

type Chan_Work struct {
	chan_work_chan_ chan *Io_Info
	once_           sync.Once
	chan_count_     int
	is_open_        bool
	events_logic_   Io_events
}

//处理链接事件
func (chan_work *Chan_Work) Do_Connect() {
	fmt.Println("")
}

func (chan_work *Chan_Work) Start(chan_count int) {
	if chan_work.is_open_ {
		return
	}

	chan_work.chan_count_ = chan_count
	chan_work.chan_work_chan_ = make(chan *Io_Info, chan_count)
	chan_work.is_open_ = true

	chan_work.events_logic_ = new(server_logic.Tcp_io_events_logic)
	chan_work.events_logic_.Init()

	//启动消费者
	go func() {
		for {
			data := <-chan_work.chan_work_chan_
			switch data.Message_type_ {
			case Io_Event_Connect:
				chan_work.events_logic_.Connect(data.Session_id_, data.Server_ip_info_, data.Client_ip_info_, data.Session_info_)
			case Io_Event_DisConnect:
				chan_work.events_logic_.Disconnect(data.Session_id_)
			case Io_Event_Data:
				chan_work.events_logic_.Recv_data(data.Session_id_, data.Mesaage_data_, data.Message_Len_, data.Session_info_)
			case Io_Exit:
				fmt.Println("[do_chan_work]summer is close(", time.Now().String(), ") do")
				return
			}
		}
	}()
}

func (chan_work *Chan_Work) Close() {
	chan_work.once_.Do(func() {
		var io_info = new(Io_Info)
		io_info.Session_id_ = 0
		io_info.Message_type_ = Io_Exit
		chan_work.chan_work_chan_ <- io_info
		close(chan_work.chan_work_chan_)
	})
}

func (chan_work *Chan_Work) Add_Message(data *Io_Info) bool {
	if len(chan_work.chan_work_chan_) >= chan_work.chan_count_ {
		//队列已经满了
		fmt.Println("[Add_Message]queue is full(", time.Now().String(), ") do")
		return false
	} else {
		chan_work.chan_work_chan_ <- data
		return true
	}

}

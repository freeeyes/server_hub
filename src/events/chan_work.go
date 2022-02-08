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
	Io_Listen_Close
	Timer_Check
)

//Io数据
type Io_session_info struct {
	Session_id_   int
	Session_info_ common.Session_Info
	active_time_  int64
}

//数据消息包
type Io_Info struct {
	Session_id_      int
	Server_ip_info_  common.Ip_info
	Client_ip_info_  common.Ip_info
	Message_type_    int
	Mesaage_data_    []byte
	Message_Len_     int
	Session_info_    common.Session_Info
	Io_LIsten_Close_ common.Io_Listen
}

type Chan_Work struct {
	chan_work_chan_  chan *Io_Info
	once_            sync.Once
	chan_count_      int
	is_open_         bool
	events_logic_    Io_events
	Io_Session_List_ map[int]*Io_session_info
	io_time_check_   int
}

//处理链接事件
func (chan_work *Chan_Work) Do_Connect() {
	fmt.Println("")
}

func (chan_work *Chan_Work) Start(chan_count int, io_timeout_millsecond int) {
	if chan_work.is_open_ {
		return
	}

	chan_work.io_time_check_ = io_timeout_millsecond
	chan_work.chan_count_ = chan_count
	chan_work.chan_work_chan_ = make(chan *Io_Info, chan_count)
	chan_work.is_open_ = true

	chan_work.events_logic_ = new(server_logic.Tcp_io_events_logic)
	chan_work.events_logic_.Init()

	//初始化map
	chan_work.Io_Session_List_ = make(map[int]*Io_session_info)

	//启动消费者
	go func() {
		for {
			data := <-chan_work.chan_work_chan_
			switch data.Message_type_ {
			case Io_Event_Connect:
				chan_work.events_logic_.Connect(data.Session_id_, data.Server_ip_info_, data.Client_ip_info_, data.Session_info_)
				//添加入当前的IO列表
				var io_session_info = new(Io_session_info)
				io_session_info.Session_id_ = data.Session_id_
				io_session_info.Session_info_ = data.Session_info_
				io_session_info.active_time_ = time.Now().Local().Unix()
				chan_work.Io_Session_List_[data.Session_id_] = io_session_info
			case Io_Event_DisConnect:
				chan_work.events_logic_.Disconnect(data.Session_id_)
				//清理入当前的IO列表
				delete(chan_work.Io_Session_List_, data.Session_id_)
			case Io_Event_Data:
				chan_work.events_logic_.Recv_data(data.Session_id_, data.Mesaage_data_, data.Message_Len_, data.Session_info_)
				if nil != chan_work.Io_Session_List_[data.Session_id_] {
					chan_work.Io_Session_List_[data.Session_id_].active_time_ = time.Now().Local().Unix()
				}
			case Io_Listen_Close:
				//关闭监听
				data.Io_LIsten_Close_.Send_finish_listen_message()
			case Timer_Check:
				//Io定时检测
				//fmt.Println("[do_chan_work]timeCheck Begin(", time.Now().String(), ") do")
				var now int64 = time.Now().Local().Unix()
				var time_interval int = chan_work.io_time_check_ / 1000
				for _, v := range chan_work.Io_Session_List_ {
					if int(now-v.active_time_) > time_interval {
						fmt.Println("[do_chan_work]timeCheck session id(", v.Session_id_, ") is timeout")
						v.Session_info_.Close_Io()
					}
				}

				//fmt.Println("[do_chan_work]timeCheck End(", time.Now().String(), ") do")
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

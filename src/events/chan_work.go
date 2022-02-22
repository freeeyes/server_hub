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
	Io_Finish
	Io_Listen_Close
	Timer_Check
)

//Io数据
type Io_session_info struct {
	Session_id_   int
	Session_info_ common.Session_Info
	active_time_  int64
}

//服务器监听接口
type Listen_io_manager interface {
	Close()
}

//IO接口管理
type Io_manager struct {
	Listen_tcp_manager_    Listen_io_manager
	Listen_udp_manager_    Listen_io_manager
	Listen_serial_manager_ Listen_io_manager
	Client_tcp_manager_    common.Client_io_manager
	Client_udp_manager_    common.Client_io_manager
}

func (io_manager *Io_manager) Set_tcp_manager(listen_tcp_manager Listen_io_manager) {
	io_manager.Listen_tcp_manager_ = listen_tcp_manager
}

func (io_manager *Io_manager) Set_udp_manager(listen_udp_manager Listen_io_manager) {
	io_manager.Listen_udp_manager_ = listen_udp_manager
}

func (io_manager *Io_manager) Set_serial_manager(listen_serial_manager Listen_io_manager) {
	io_manager.Listen_serial_manager_ = listen_serial_manager
}

func (io_manager *Io_manager) Set_client_tcp_manager(listen_client_tcp_manager common.Client_io_manager) {
	io_manager.Client_tcp_manager_ = listen_client_tcp_manager
}

func (io_manager *Io_manager) Set_client_udp_manager(listen_client_udp_manager common.Client_io_manager) {
	io_manager.Client_udp_manager_ = listen_client_udp_manager
}

//数据消息包
type Io_Info struct {
	Session_id_      int
	Message_type_    int
	Mesaage_data_    []byte
	Message_Len_     uint32
	Session_info_    common.Session_Info
	Io_LIsten_Close_ common.Io_Listen
	Command_id_      uint16
}

type Chan_Work struct {
	chan_work_chan_        chan *Io_Info
	once_                  sync.Once
	chan_count_            int
	is_open_               bool
	Io_Session_List_       map[int]*Io_session_info
	io_time_check_         int
	logic_command_list_    map[uint16]func(int, []byte, int, common.Session_Info)
	logic_list_            []common.Server_logic_info
	client_tcp_manager_    common.Client_io_manager
	client_udp_manager_    common.Client_io_manager
	listen_tcp_manager_    Listen_io_manager
	listen_udp_manager_    Listen_io_manager
	listen_serial_manager_ Listen_io_manager
}

func (chan_work *Chan_Work) Get_tcp_clientmanager() common.Client_io_manager {
	return chan_work.client_tcp_manager_
}

func (chan_work *Chan_Work) Get_udp_clientmanager() common.Client_io_manager {
	return chan_work.client_udp_manager_
}

func (chan_work *Chan_Work) Add_listen_tcp_manager(listen_tcp_manager Listen_io_manager) {
	chan_work.listen_tcp_manager_ = listen_tcp_manager
}

func (chan_work *Chan_Work) Add_listen_udp_manager(listen_udp_manager Listen_io_manager) {
	chan_work.listen_udp_manager_ = listen_udp_manager
}

func (chan_work *Chan_Work) Add_listen_serial_manager(listen_serial_manager Listen_io_manager) {
	chan_work.listen_serial_manager_ = listen_serial_manager
}

//注册tcp服务器间对象
func (chan_work *Chan_Work) Add_tcp_client_manager(client_tcp_manager common.Client_io_manager) {
	chan_work.client_tcp_manager_ = client_tcp_manager
}

//注册udp服务器间对象
func (chan_work *Chan_Work) Add_udp_client_manager(client_udp_manager common.Client_io_manager) {
	chan_work.client_udp_manager_ = client_udp_manager
}

//注册消费逻辑
func (chan_work *Chan_Work) Regedit_command(command_id uint16, logic_func func(int, []byte, int, common.Session_Info)) bool {
	fmt.Println("[Regedit_command]command_id=", command_id)
	if _, ok := chan_work.logic_command_list_[command_id]; !ok {
		//没有找到，添加
		chan_work.logic_command_list_[command_id] = logic_func
		return true
	} else {
		return false
	}
}

//处理消费逻辑
func (chan_work *Chan_Work) Get_logic(command_id uint16) (bool, func(int, []byte, int, common.Session_Info)) {
	if logic_func, ok := chan_work.logic_command_list_[command_id]; ok {
		//找到，返回
		return true, logic_func
	} else {
		return false, nil
	}
}

func (chan_work *Chan_Work) do_connect(data *Io_Info) {
	if ret, logic_func := chan_work.Get_logic(uint16(Io_Event_Connect)); ret {
		//找到了对应服务，执行
		logic_func(data.Session_id_, nil, 0, data.Session_info_)
	}

	//添加入当前的IO列表
	var io_session_info = new(Io_session_info)
	io_session_info.Session_id_ = data.Session_id_
	io_session_info.Session_info_ = data.Session_info_
	io_session_info.active_time_ = time.Now().Local().Unix()
	chan_work.Io_Session_List_[data.Session_id_] = io_session_info
}

func (chan_work *Chan_Work) do_disconnect(data *Io_Info) {
	if ret, logic_func := chan_work.Get_logic(uint16(Io_Event_DisConnect)); ret {
		//找到了对应服务，执行
		logic_func(data.Session_id_, nil, 0, data.Session_info_)
	}
	//清理入当前的IO列表
	delete(chan_work.Io_Session_List_, data.Session_id_)
}

func (chan_work *Chan_Work) do_logic(data *Io_Info) {
	if ret, logic_func := chan_work.Get_logic(data.Command_id_); ret {
		//找到了对应服务，执行
		logic_func(data.Session_id_, data.Mesaage_data_, int(data.Message_Len_), data.Session_info_)
	}
	if nil != chan_work.Io_Session_List_[data.Session_id_] {
		chan_work.Io_Session_List_[data.Session_id_].active_time_ = time.Now().Local().Unix()
	}
}

func (chan_work *Chan_Work) do_close_listen(data *Io_Info) {
	//关闭监听
	data.Io_LIsten_Close_.Send_finish_listen_message()
}

func (chan_work *Chan_Work) do_time_check() {
	//Io定时检测
	var now int64 = time.Now().Local().Unix()
	var time_interval int = chan_work.io_time_check_ / 1000
	for _, v := range chan_work.Io_Session_List_ {
		if int(now-v.active_time_) > time_interval {
			fmt.Println("[do_chan_work]timeCheck session id(", v.Session_id_, ") is timeout")
			v.Session_info_.Close_Io()
		}
	}

	//fmt.Println("[do_chan_work]timeCheck End(", time.Now().String(), ") do")
	//服务器间链接定时监测
	if chan_work.client_tcp_manager_ != nil {
		chan_work.client_tcp_manager_.Time_Check()
	}

	if chan_work.client_udp_manager_ != nil {
		chan_work.client_udp_manager_.Time_Check()
	}
}

func (chan_work *Chan_Work) Run() {
	//启动消费者
	for {
		data := <-chan_work.chan_work_chan_
		switch data.Message_type_ {
		case Io_Event_Connect:
			chan_work.do_connect(data)
		case Io_Event_DisConnect:
			chan_work.do_disconnect(data)
		case Io_Event_Data:
			chan_work.do_logic(data)
		case Io_Listen_Close:
			chan_work.do_close_listen(data)
		case Timer_Check:
			chan_work.do_time_check()
		case Io_Finish:
			//全部关闭完成
			fmt.Println("[do_chan_work]summer is close(", time.Now().String(), ") do")
			chan_work.Close()
			return
		}
	}
}

func (chan_work *Chan_Work) Start(chan_count int, io_timeout_millsecond int) {
	if chan_work.is_open_ {
		return
	}

	chan_work.io_time_check_ = io_timeout_millsecond
	chan_work.chan_count_ = chan_count
	chan_work.chan_work_chan_ = make(chan *Io_Info, chan_count)
	chan_work.is_open_ = true

	//初始化映射表
	chan_work.logic_command_list_ = make(map[uint16]func(int, []byte, int, common.Session_Info))

	//初始化加载模块列表
	chan_work.logic_list_ = make([]common.Server_logic_info, 10)

	//初始化map
	chan_work.Io_Session_List_ = make(map[int]*Io_session_info)

	//加载初始化模块
	var server_logic_info common.Server_logic_info = new(server_logic.Tcp_io_events_logic)
	server_logic_info.Init(chan_work)
	chan_work.logic_list_ = append(chan_work.logic_list_, server_logic_info)
}

func (chan_work *Chan_Work) Close() {
	chan_work.once_.Do(func() {
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

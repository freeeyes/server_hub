package socket

import (
	"fmt"
	"net"
	"server_hub/events"
	"strconv"
	"strings"
)

type Tcp_client struct {
	session_id_     int
	session_        *Tcp_Session
	client_ip_      string
	client_port_    string
	server_ip_      string
	server_port_    string
	client_conn_    net.Conn
	chan_work_      *events.Chan_Work
	recv_buff_size_ int
	send_buff_size_ int
	packet_parse_   events.Io_buff_to_packet
}

func (tcp_client *Tcp_client) Close() {
	//发送链接断开事件
	var message = new(events.Io_Info)
	message.Session_id_ = tcp_client.session_.Get_Session_ID()
	message.Message_type_ = events.Io_Event_DisConnect
	tcp_client.chan_work_.Add_Message(message)

	tcp_client.client_conn_.Close() // 关闭TCP连接
}

func (tcp_client *Tcp_client) Connct(session_id int, server_ip string, server_port int, recv_buff_size int, send_buff_size int, packet_parse events.Io_buff_to_packet, chan_work *events.Chan_Work) bool {
	tcp_client.session_id_ = session_id
	tcp_client.server_ip_ = server_ip
	tcp_client.server_port_ = strconv.Itoa(server_port)
	tcp_client.recv_buff_size_ = recv_buff_size
	tcp_client.send_buff_size_ = send_buff_size
	tcp_client.packet_parse_ = packet_parse
	tcp_client.chan_work_ = chan_work

	conn, err := net.Dial("tcp", tcp_client.server_ip_+":"+tcp_client.server_port_)
	if err != nil {
		fmt.Println("[tcp_client::connct]connect error=", err)
		return false
	}

	tcp_client.client_conn_ = conn
	defer tcp_client.Close()

	tcp_client.client_ip_ = strings.Split(conn.LocalAddr().String(), ":")[0]
	tcp_client.client_port_ = strings.Split(conn.LocalAddr().String(), ":")[1]

	tcp_client.session_ = new(Tcp_Session)
	tcp_client.session_.Init(session_id,
		tcp_client.client_ip_,
		tcp_client.client_port_,
		server_ip,
		tcp_client.server_port_,
		tcp_client.client_conn_,
		tcp_client.recv_buff_size_,
		tcp_client.send_buff_size_)

	for {
		reqLen, err := tcp_client.client_conn_.Read(tcp_client.session_.Get_recv_buff())
		if err != nil || reqLen == 0 {
			fmt.Println("[Tcp_client::Handle_Connection]session id=", tcp_client.session_.Get_Session_ID(), ",err=", err)
			break
		}

		fmt.Println("[Tcp_client::Handle_Connection]session id=", tcp_client.session_.Get_Session_ID(), " is recv, datalen=", reqLen)
		tcp_client.session_.Set_write_len(reqLen)

		//在这里数据包拆包分析
		packet_list, read_len, parse_is_ok := packet_parse.Recv_buff_to_packet(tcp_client.session_.Get_read_buff(), tcp_client.session_.Get_write_buff())
		//fmt.Println("[Handle_Connection]packet_list=", len(packet_list))
		//fmt.Println("[Handle_Connection]read_len=", read_len)
		if !parse_is_ok {
			fmt.Println("[Tcp_client::Handle_Connection]parse_is_ok=", parse_is_ok)
			break
		}

		for _, packet := range packet_list {
			var message = new(events.Io_Info)
			message.Session_id_ = tcp_client.session_.Get_Session_ID()
			message.Message_type_ = events.Io_Event_Data
			message.Mesaage_data_ = packet.Data_
			message.Message_Len_ = packet.Data_len_
			message.Command_id_ = packet.Command_id_
			message.Session_info_ = tcp_client.session_
			tcp_client.chan_work_.Add_Message(message)
		}

		tcp_client.session_.Reset_read_buff(read_len)
		fmt.Println("[Tcp_client::Handle_Connection]writelen=", tcp_client.session_.Get_write_buff())
	}

	return true
}

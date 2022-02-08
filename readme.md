Table of Contents
=================

 - [Overview](#overview)
 - [Build and Install](#build-and-install)
 - [how to config](#[how-to-config)
 - [how to use](#how-to-use)

Overview
========
This is an integrated IO link sub server.
Support tcp udp and serial protocol.  

Build and Install
=================
 * go build  

how to config
==========
open file config/server_config.json
```json  
{
    "Tcp_server_":[{
        "Server_ip_":"127.0.0.1",
        "Server_port_":"10020"
    }],

     "Udp_Server_":[{
      "Server_ip_":"127.0.0.1",
      "Server_port_":"10030"
    }],

    "Serial_Server_":[{
    }],

    "Recv_queue_count_":100,
    "Recv_buff_size_":1024,
    "Send_buff_size_":1024,
    "Io_time_check_":10000
  }
```

You can enable multiple tcp, udp, serial ports, monitoring in the configuration file at the same time.  
"Io_time_check_" is can check your Io timeout, the unit is milliseconds.  
if you don't need this feature, you can set 0.  

"Recv_buff_size_" is IO recv packet buffer max size. 
"Send_buff_size_" is IO send packet buffer max size.  
"Recv_queue_count_" is work chan max size.  

how to use
==========
you can write file write file "io_logic.go"  
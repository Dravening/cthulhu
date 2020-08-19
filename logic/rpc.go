package logic

import (
	"cthulhu/utils"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type Cthulhu struct {
}

func (cthulhu Cthulhu) MultiEstablish(sshInfo utils.SSHInfo, reply *[]string) error {
	device := NewDeviceSSH(sshInfo)
	*reply = MultiEstablish(device)
	return nil
}

func StartServer() {
	rpc.Register(new(Cthulhu))
	rpc.HandleHTTP()
	//http.ListenAndServe("127.0.0.1:8081", nil)
	tcpCon, err := net.Listen("tcp", "127.0.0.1:8880")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	http.Serve(tcpCon, nil)
}

func StartClient() {
	sshInfo := utils.SSHInfo{
		AddrStr:      "192.168.2.200:22;192.168.2.201:22",
		Cmd:          "show clock;config t;do show run;exit",
		User:         "test",
		Password:     "test",
		Ciphers:      "",
		Key:          "",
		Timeout:      30,
		RoutineLimit: 20,
	}
	client, err := rpc.DialHTTP("tcp", "127.0.0.1"+":8880")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	replyList := []string{}
	err = client.Call("Cthulhu.MultiEstablish", sshInfo, &replyList)
	if err != nil {
		log.Fatal("error:", err)
	}
	//you may use the replyList in another way.
	for i := 0; i < len(replyList); i++ {
		fmt.Println(replyList[i])
	}
}

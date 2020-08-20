# cthulhu
Cthulhu is an automatic tool which can easily manage network devices.

#### functions
- Concurrent execution(connect with numbers of devices at the same time).
- run Multiple commands in one order
- ssh auth in user/password way
- ssh auth in key way

#### usages
Cthulhu could run in two modes -- cmd mode && rpc mode.
- cmd mode:
```go
./cthulhu -C "show clock;config t;do show run;exit" -A "192.168.1.1:22;192.168.1.2:22" -U test -P test
```

- rpc server:
Cthulhu could be your rpc server,providing the "manage multi-devices" service.
```go
./cthulhu serve
```
- rpc client:
Then you need to implement an rpc-client yourself. For example:
```go
func StartClient() {
	sshInfo := utils.SSHInfo{
		AddrStr:      "192.168.1.1:22;192.168.1.2:22",
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
```

You can type `./cthulhu h` for more information.

#### LICENSE
Apache License 2.0
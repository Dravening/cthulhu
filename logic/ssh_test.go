package logic

import (
	"cthulhu/utils"
	"strings"
	"testing"
)

func TestMultiEstablish(t *testing.T) {
	//-------your environment for test--------
	address := "192.168.2.201:22"
	cmd := "config t;exit"
	user := "test"
	password := "test"
	//----------------------------------------

	sshInfo := utils.SSHInfo{
		AddrStr:      address,
		Cmd:          cmd,
		User:         user,
		Password:     password,
		Ciphers:      "",
		Key:          "",
		Timeout:      30,
		RoutineLimit: 30,
	}
	device := NewDeviceSSH(sshInfo)
	replyList := MultiEstablish(device)

	num := len(replyList) - 1
	if strings.Contains(replyList[num], "#exit") {
		return
	}
	t.Error("MultiEstablish unit test fail.")
}

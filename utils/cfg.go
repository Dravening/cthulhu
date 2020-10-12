package utils

import (
	"strings"
)

func SplitString(str string) (strList []string) {
	if strings.Contains(str, ",") {
		strList = strings.Split(str, ",")
	} else {
		strList = strings.Split(str, ";")
	}
	return
}

type SSHInfo struct {
	AddrStr      string
	Cmd          string
	User         string
	Password     string
	Ciphers      string
	Key          string
	Protocol     string
	Timeout      int64
	RoutineLimit int
}

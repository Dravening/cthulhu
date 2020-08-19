package logic

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"time"

	"cthulhu/utils"

	"golang.org/x/crypto/ssh"
)

type DeviceSSH struct {
	CMDs         []string
	Addr         string
	Key          string
	Password     string
	Ciphers      string
	Timeout      int64
	ClientConfig *ssh.ClientConfig
	RoutineConf  RoutineConf
}

type RoutineConf struct {
	RoutineLimit int
	AddrList     []string
}

type ConnResult struct {
	Success bool
	Addr    string
	Result  string
}

func NewDeviceSSH(sshInfo utils.SSHInfo) DeviceSSH {
	device := DeviceSSH{
		CMDs: utils.SplitString(sshInfo.Cmd),
		ClientConfig: &ssh.ClientConfig{
			Auth: make([]ssh.AuthMethod, 0),
		},
	}
	device.ClientConfig.User = sshInfo.User
	device.Timeout = sshInfo.Timeout
	device.Ciphers = sshInfo.Ciphers
	device.Key = sshInfo.Key
	device.Password = sshInfo.Password
	device.RoutineConf.AddrList = utils.SplitString(sshInfo.AddrStr)
	device.RoutineConf.RoutineLimit = sshInfo.RoutineLimit
	return device
}

func MultiEstablish(device DeviceSSH) []string {
	var reply string
	var replyList []string
	resultsChan := make([]chan ConnResult, len(device.RoutineConf.AddrList))
	routineLimit := make(chan bool, device.RoutineConf.RoutineLimit)

	routine := func(routineLimit chan bool, device DeviceSSH, ch chan ConnResult) {
		Establish(device, ch)
		<-routineLimit
	}
	for i, addr := range device.RoutineConf.AddrList {
		resultsChan[i] = make(chan ConnResult, 1)
		device.Addr = addr
		routineLimit <- true
		go routine(routineLimit, device, resultsChan[i])
	}
	//results := []ConnResult{}
	for _, ch := range resultsChan {
		result := <-ch
		reply = fmt.Sprintf(`address: %s
---------- Conversation ----------
%s`, result.Addr, result.Result)
		fmt.Println(reply)
		replyList = append(replyList, reply)
	}
	return replyList
}

func DialConn(device DeviceSSH) (*ssh.Session, error) {
	var err error
	var session *ssh.Session
	// clientConfig.Auth
	auth := make([]ssh.AuthMethod, 0)
	if device.Key != "" {
		keyBytes, err := ioutil.ReadFile(device.Key)
		if err != nil {
			return nil, err
		}
		var signer ssh.Signer
		if device.Password == "" {
			signer, err = ssh.ParsePrivateKey(keyBytes)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(device.Password))
		}
		if err != nil {
			return nil, err
		}
		auth = append(device.ClientConfig.Auth, ssh.PublicKeys(signer))
	} else {
		auth = append(device.ClientConfig.Auth, ssh.Password(device.Password))
	}
	device.ClientConfig.Auth = auth

	// clientConfig.Config
	device.ClientConfig.Config = ssh.Config{
		Ciphers:      []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		KeyExchanges: []string{"diffie-hellman-group-exchange-sha1", "diffie-hellman-group1-sha1", "diffie-hellman-group-exchange-sha256"},
	}
	if device.Ciphers != "" {
		ciphersList := utils.SplitString(device.Ciphers)
		device.ClientConfig.Config = ssh.Config{
			Ciphers: ciphersList,
		}
	}

	// clientConfig.Timeout.
	//If the time has run out before the connection is complete, an error is returned.
	//Once successfully connected, any expiration of the context will not affect the connection.
	if device.ClientConfig.Timeout == 0 {
		device.ClientConfig.Timeout = 30 * time.Second
	}

	// clientConfig.HostKeyCallback
	device.ClientConfig.HostKeyCallback = func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }

	// connect  device
	client, err := ssh.Dial("tcp", device.Addr, device.ClientConfig)
	if err != nil {
		return nil, err
	}
	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}
	if err := session.RequestPty("vt100", 238, 2000, modes); err != nil {
		return nil, err
	}
	return session, nil
}

func Establish(device DeviceSSH, ch chan ConnResult) {
	connResult := ConnResult{}
	chSession := make(chan ConnResult)

	go session(device, chSession)
	select {
	case connResult = <-chSession:
		ch <- connResult
	case <-time.After(time.Duration(device.Timeout) * time.Second):
		connResult.Success = false
		connResult.Addr = device.Addr
		connResult.Result = "establish session failed,timeout(" + strconv.FormatInt(device.Timeout, 10) + "s)."
		ch <- connResult
	}
	return
}

func session(device DeviceSSH, ch chan ConnResult) {
	var (
		conn      ConnResult
		outBuffer bytes.Buffer
		errBuffer bytes.Buffer
	)
	conn.Addr = device.Addr
	cmds := append(device.CMDs, "exit")

	session, err := DialConn(device)
	if err != nil {
		conn.Success = false
		conn.Result = err.Error()
		ch <- conn
		return
	}
	session.Stdout = &outBuffer
	session.Stderr = &errBuffer
	stdinBuffer, _ := session.StdinPipe()
	defer session.Close()

	err = session.Shell()
	if err != nil {
		conn.Success = false
		conn.Result = err.Error()
		ch <- conn
		return
	}
	for _, c := range cmds {
		c = c + "\n"
		stdinBuffer.Write([]byte(c))
	}
	session.Wait()
	if errBuffer.String() != "" {
		conn.Success = false
		conn.Result = errBuffer.String()
		ch <- conn
	} else {
		conn.Success = true
		conn.Result = outBuffer.String()
		ch <- conn
	}
	return
}

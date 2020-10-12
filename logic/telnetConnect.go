package logic

import (
	"log"
	"net"
	"strings"
	"time"
)

type Device struct {
	CMDs      []string
	Addr      string
	Timeout   int64
	Protocol  string
	TLSConfig tlsConfig
}

type tlsConfig struct {
	user   string
	passwd string
}

const (
	TIME_DELAY_AFTER_WRITE = 1000 //1000ms
)

func sessionTELNET(device Device, ch chan ConnResult) {
	var (
		connRst   ConnResult
		outBuffer [4096]byte
	)

	conn, err := device.telnet()
	if err != nil {
		connRst.Addr = device.Addr
		connRst.Success = false
		connRst.Result = err.Error()
		ch <- connRst
		return
	}
	defer conn.Close()

	conn.Write([]byte("\n\r"))
	time.Sleep(time.Millisecond * TIME_DELAY_AFTER_WRITE)
	n, _ := conn.Read(outBuffer[:])

	connRst.Result = connRst.Result + string(outBuffer[14:n])
	outBuffer = [4096]byte{}

	//get device name
	conn.Write([]byte("\n\r"))
	time.Sleep(time.Millisecond * TIME_DELAY_AFTER_WRITE)
	n, _ = conn.Read(outBuffer[:])
	deviceName := string(outBuffer[2 : n-2])
	outBuffer = [4096]byte{}

	for _, c := range device.CMDs {
		c = c + "\n\r"
		conn.Write([]byte(c))

		for {

			time.Sleep(time.Millisecond * TIME_DELAY_AFTER_WRITE)
			n, _ := conn.Read(outBuffer[:])

			if strings.Contains(string(outBuffer[:]), "hostname") {
				connRst.Result = connRst.Result + string(outBuffer[:n])
				if strings.Count(string(outBuffer[:]), deviceName) >= 2 {
					outBuffer = [4096]byte{}
					break
				}
				outBuffer = [4096]byte{}
			} else if strings.Contains(string(outBuffer[:]), deviceName) {
				connRst.Result = connRst.Result + string(outBuffer[:n])
				outBuffer = [4096]byte{}
				break
			} else {
				connRst.Result = connRst.Result + string(outBuffer[:n])
				outBuffer = [4096]byte{}
			}

		}
	}

	connRst.Addr = device.Addr
	connRst.Success = true
	//connRst.Result = string(outBuffer[:])
	ch <- connRst
}

func (device *Device) telnet() (net.Conn, error) {
	addr := device.Addr
	conn, err := net.DialTimeout("tcp", addr, time.Duration(device.Timeout)*time.Second)
	if nil != err {
		log.Fatalln("pkg: model, func: Telnet, method: net.DialTimeout, errInfo:", err)
		return nil, err
	}

	if false == device.telnetProtocolHandshake(conn) {
		log.Fatalln("pkg: model, func: Telnet, method: this.telnetProtocolHandshake, errInfo: telnet protocol handshake failed!!!")
		return nil, err
	}
	return conn, nil
}

func (device *Device) telnetProtocolHandshake(conn net.Conn) bool {
	var wBuf, rBuf [4096]byte
	_, err := conn.Read(rBuf[0:])
	if nil != err {
		log.Fatalln("pkg: model, func: telnetProtocolHandshake, method: conn.Read, errInfo:", err)
		return false
	}

	wBuf[1] = 252
	wBuf[4] = 252
	wBuf[7] = 252
	wBuf[10] = 252
	_, err = conn.Write(wBuf[:])
	if nil != err {
		log.Fatalln("pkg: model, func: telnetProtocolHandshake, method: conn.Write, errInfo:", err)
		return false
	}
	time.Sleep(time.Millisecond * TIME_DELAY_AFTER_WRITE)

	wBuf[1] = 252
	wBuf[4] = 251
	wBuf[7] = 252
	wBuf[10] = 254
	wBuf[13] = 252
	_, err = conn.Write(wBuf[:])
	if nil != err {
		log.Fatalln("pkg: model, func: telnetProtocolHandshake, method: conn.Write, errInfo:", err)
		return false
	}
	time.Sleep(time.Millisecond * 300)

	if "" == device.TLSConfig.user {
		return true
	}

	_, err = conn.Write([]byte(device.TLSConfig.user + "\n"))
	if nil != err {
		log.Fatalln("pkg: model, func: telnetProtocolHandshake, method: conn.Write, errInfo:", err)
		return false
	}
	time.Sleep(time.Millisecond * TIME_DELAY_AFTER_WRITE)

	_, err = conn.Write([]byte(device.TLSConfig.passwd + "\n"))
	if nil != err {
		log.Fatalln("pkg: model, func: telnetProtocolHandshake, method: conn.Write, errInfo:", err)
		return false
	}
	time.Sleep(time.Millisecond * TIME_DELAY_AFTER_WRITE)

	return true
}

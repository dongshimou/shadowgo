package test

import (
	"io"
	"log"
	"net"
	"strconv"
)

var lconn net.Conn
var rconn net.Conn

func TcpProxy() {
	l, err := net.Listen("tcp", ":1080")
	if err != nil {
		log.Println(" listen error")
		return
	}
	log.Println("start listen")
	for {
		lconn, err = l.Accept()
		log.Println("start a accept")
		if err != nil {
			log.Println("accept error")
			continue
		}
		defer lconn.Close()

		b := make([]byte, 2048)

		_, err = lconn.Read(b)

		logSocks := func(j int) {
			for i := 0; i < j; i++ {
				print(b[i], " ")
			}
			print("\n")
		}

		//logSocks(0)
		logSocks(3)

		if b[0] == 0x05 {
			switch b[2] {
			case 0x00: //无密码
				lconn.Write([]byte{0x05, 0x00})
			case 0x01: //通用安全接口
				lconn.Write([]byte{0x05, 0x01})
			case 0x02: //用户名+密码
				lconn.Write([]byte{0x05, 0x02})
				// 0x03 ~ 0x7F IANA分配
				// 0x80 ~ 0xFE 私人方法
			case 0xFF: //无方法
				log.Println("no method")
				continue
			}
		} else {
			log.Println("it is not socksV5 ")
			continue
		}

		var n int
		n, err = lconn.Read(b)

		logSocks(10)

		if b[0] == 0x05 {
			//0x01 connect
			//0x02 bind
			//0x03 udp associate
			if b[1] == 0x01 {
				var host, port string
				switch b[3] {
				case 0x01: //ipv4
					host = net.IP{b[4], b[5], b[6], b[7]}.String()
				case 0x03: //host
					host = string(b[5 : n-2])
				case 0x04: //ipv6
					host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
				}
				port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
				println(host, port)

				rconn, err := net.Dial("tcp", net.JoinHostPort(host, port))
				if err != nil {
					log.Println("join error")
					continue
				}
				defer rconn.Close()

				lconn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})

				go io.Copy(rconn, lconn)
				io.Copy(lconn, rconn)
			}
		}
	}
}

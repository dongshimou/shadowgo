package src

import (
	"../../common"
	"encoding/binary"
	"log"
	"net"
	"strconv"
)

type tcpMsg struct {
	id  uint64
	buf []byte
}

var cpool = make(map[uint64]tcpMsg)
var uid uint64 = 0
var sendqueue chan tcpMsg
var recvqueue chan tcpMsg

func Listen() {

	l, err := net.Listen("tcp", ":12344")
	if err != nil {
		log.Println(" listen error")
		return
	}

	sendqueue = make(chan tcpMsg, 4096)
	recvqueue = make(chan tcpMsg, 4096)

	log.Println("start listen")
	for {
		lconn, err := l.Accept()
		log.Println("start a accept")
		if err != nil {
			log.Println("accept error")
			return
		}
		go sockv5(lconn)
	}
}

func send(c chan tcpMsg, rconn net.Conn) {
	for {
		res := <-c
		bef := make([]byte, 16)
		binary.BigEndian.PutUint64(bef, res.id)
		bef = append(bef[:16], res.buf...)
		rconn.Write(bef)
	}
}

func recv(c chan tcpMsg, rconn net.Conn) {
	for {
		buf := make([]byte, 32*1024)
		for {
			_, err := rconn.Read(buf)
			if err != nil {

			}
			msg := tcpMsg{
				id:  binary.BigEndian.Uint64(buf[:16]),
				buf: buf[16:],
			}
			c <- msg
		}
	}
}

func sockv5(lconn net.Conn) {

	defer lconn.Close()
	b := make([]byte, 2048)

	/* request
	-------------------------------------
	|	VER		NMETHODS	METHODS		|
	|	1byte	1byte		1byte		|
	-------------------------------------
	*/

	n, err := lconn.Read(b)
	if err != nil {
		return
	}
	DEBUG := false
	logSocks := func(j int) {
		if !DEBUG {
			return
		}
		for i := 0; i < j; i++ {
			print(b[i], " ")
		}
		print("\n")
	}

	/* response
	-------------------------
	|	VER		METHOD		|
	|	1byte	1byte		|
	-------------------------
	*/

	//logSocks(0)
	logSocks(n)

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
			return
		}
	} else {
		log.Println("it is not socksV5 ")
		return
	}
	n, err = lconn.Read(b)

	logSocks(n)

	/* request
	-------------------------------------------------------------
	|	VER		CMD		RSV		ATYP	DST.ADDR	DST.PORT	|
	|	1byte	1byte	1byte	1byte	n byte		2byte		|
	-------------------------------------------------------------
	*/

	if b[0] == 0x05 {
		//解析主机和端口
		parseHostPort := func(b []byte, n int) (host, port string) {
			switch b[3] {
			case 0x01: //ipv4	4字节
				host = net.IP{b[4], b[5], b[6], b[7]}.String()
			case 0x03: //host	字符串
				host = string(b[5 : n-2])
			case 0x04: //ipv6	16字节
				addr := b[4:20]
				host = net.IP(addr).String()
			}
			//端口	2字节
			port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
			println(host, port)
			return host, port
		}

		host, port := parseHostPort(b, n)
		switch b[1] {
		case 0x01: //0x01 connect

			/* response
			-------------------------------------------------------------
			|	VER		REP		RSV		ATYP	DST.ADDR	DST.PORT	|
			|	1byte	1byte	1byte	1byte	n byte		2byte		|
			-------------------------------------------------------------
			*/

			res := make([]byte, 10)

			// 版本号
			res[0] = 0x05
			// 返回状态

			//	0x00	成功
			//	0x01	普通的失败
			//	0x02	规则不允许的连接
			//	0x03	网络不可达
			//	0x04	主机不可达
			//	0x05	连接被拒绝
			//	0x06	TTL超时
			//	0x07	不支持的命令
			//	0x08	不支持的地址类型
			//	0x09 ~ 0xFF		未定义
			res[1] = 0x00

			//	保留位
			res[2] = 0x00

			//	地址类型
			//	0x01	IPV4
			//	0x03	域名
			//	0x04	IPV6
			res[3] = 0x01

			// IPV4为4字节 域名为字符串 IPV6为16字节
			res[4], res[5], res[6], res[7] = 0x00, 0x00, 0x00, 0x00

			//	端口为2字节
			res[8], res[9] = 0x00, 0x00

			lconn.Write(res)
			//hello(lconn, host, port)
			tcpProxy(lconn, host, port)
		case 0x02: //0x02 bind
			//尚不清楚，貌似是udp的特殊绑定
		case 0x03: //0x03 udp associate
			udpProxy(lconn, host, port)
		}
	}
}

//test function
func hello(lconn net.Conn, host, port string) {
	///todo
	//目前是直接与host&port建立的tcp连接
	rconn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Println("join error")
		return
	}
	defer rconn.Close()

	///todo
	//转发

	//测试代码 打印http请求
	//b:=make([]byte,4096)
	//_, err = lconn.Read(b)
	//println(string(b))
	//rconn.Write(b)

	go common.CopyData(rconn, lconn)
	log.Println("client -> server")
	common.CopyData(lconn, rconn)
	log.Println("server -> client")

	//time.Sleep(time.Second*60)
	log.Println("req && res copy over")
}

func udpProxy(lconn net.Conn, host, port string) {

	rconn, err := net.Dial("udp", net.JoinHostPort(host, port))
	if err != nil {
		log.Println("join error")
		return
	}
	defer rconn.Close()

	/* request
	-------------------------------------------------------------
	|	RSV		FRAG	ATYP	DST.ADDR	DST.PORT	DATA	|
	|	2byte	1byte	1byte	n byte		2byte		n byte	|
	-------------------------------------------------------------
	*/

	///todo
	//udp 转发
}
var SQUEUE = make(chan []byte, 512)
var RQUEUE = make(chan []byte,512)

func initProxy(host,port string){
	channal, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err!=nil{
		log.Println(err)
		panic(err)
	}
	SQUEUE<-[]byte("2333")


	send:= func() {
		for {
			data := <-SQUEUE
			channal.Write(data)
		}
	}
	recv:= func() {
		for {
			msg := make([]byte, 4096)
				_, err := channal.Read(msg)
				if err!=nil{
					log.Println(err)
					break
				}
			}
		}
	go send()
	go recv()
}

func tcpProxy(lconn net.Conn, host, port string) {

	//与远端sg服务器建立链接
	log.Println("start connet remote server !")
	serhost, serport := "127.0.0.1", "23333"

	rconn, err := net.Dial("tcp", net.JoinHostPort(serhost, serport))

	if err != nil {
		log.Println("connect server error !")
		return
	}
	defer rconn.Close()

	//本地与远端 shadowgo的握手包
	//todo 加密解密数据
	//转发需要代理的host&&port
	//转发数据

	b := make([]byte, 1024)

	/* request
	-------------------------------------
	|	VER		NULL		METHODS		|
	|	1byte	1byte		1byte		|
	-------------------------------------
	*/

	b[0], b[1], b[2] = 0xDD, 0x00, 0x01

	rconn.Write(b)

	/* request
	-------------------------------------
	|	VER		STATUS		METHODS		|
	|	1byte	1byte		1byte		|
	-------------------------------------
	*/

	_, err = rconn.Read(b)

	if b[0] == 0xDD {
		if b[1] != 0x00 {
			log.Println("ack 0x00 error")
			return
		}
		switch b[2] {
		case 0x01: //成功
			{
				/* response
				|	VER		STATUS	HLEN	HOST	PORT
				|	1byte	1byte	2byte	n byte	2byte
				*/
				buffer := make([]byte, 1024)
				buffer[0] = 0xDD
				buffer[1] = 0x01
				hb := []byte(host)

				string2blen := func(port string) []byte {
					p64, _ := strconv.ParseInt(port, 10, 32)
					p := int(p64)
					base := 1 << 8
					res := make([]byte, 2)
					res[0] = byte(p / base)
					res[1] = byte(p % base)
					return res
				}
				hostlen := string2blen(strconv.Itoa(len(host)))
				buffer = append(buffer[:2], hostlen[0], hostlen[1])
				buffer = append(buffer[:4], hb...)
				portlen := string2blen(port)
				buffer = append(buffer, portlen[0], portlen[1], 0x00)
				log.Println("=============", host, port)
				rconn.Write(buffer)
				{

					log.Println("====start copy====")
					go common.CopyData(rconn, lconn)
					common.CopyData(lconn, rconn)

					log.Println("====finish====")
					if b[0] == 0xdd {
						if b[1] == 0x02 {
							if b[2] == 0xdd {
								return
							}
						}
					}
				}
			}
		case 0x02: //一般性失败

		case 0x03: //不支持的加密
		//b[2]为服务端的加密方式
		case 0x04: //账号密码错误

		case 0x05: //超时
		}
	} else {
		log.Println("0xdd error")
		return
	}
}

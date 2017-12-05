package test

import (
	"io"
	"log"
	"net"
	"strconv"
)

func SocksProxy() {
	l, err := net.Listen("tcp", ":1080")
	if err != nil {
		log.Println(" listen error")
		return
	}
	log.Println("start listen")
	for {
		lconn, err := l.Accept()
		log.Println("start a accept")
		if err != nil {
			log.Println("accept error")
			return
		}
		defer lconn.Close()
		go startProxy(lconn)
	}
}

func startProxy(lconn net.Conn) {

	b := make([]byte, 2048)

	/* request
	-------------------------------------
	|	VER		NMETHODS	METHODS		|
	|	1byte	1byte		1byte		|
	-------------------------------------
	*/

	n, err := lconn.Read(b)
	if err != nil {

	}
	logSocks := func(j int) {
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

	logSocks(0)
	//logSocks(n)
	//println(string(b))

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

	//logSocks(10)

	/* request
	-------------------------------------------------------------
	|	VER		CMD		RSV		ATYP	DST.ADDR	DST.PORT	|
	|	1byte	1byte	1byte	1byte	n byte		2byte		|
	-------------------------------------------------------------
	*/

	if b[0] == 0x05 {

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
			tcpProxy(lconn, host, port)
		case 0x02: //0x02 bind

		case 0x03: //0x03 udp associate
			udpProxy(lconn, host, port)
		}
	}
}
func tcpProxy(lconn net.Conn, host, port string) {
	///todo
	//目前是直接与host&port建立的tcp连接
	rconn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		log.Println("join error")
		return
	}
	defer rconn.Close()

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

	///todo
	//转发

	//测试代码 打印http请求
	//n, err = lconn.Read(b)
	//println(string(b))
	//rconn.Write(b)

	go io.Copy(rconn, lconn)

	io.Copy(lconn, rconn)

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
}

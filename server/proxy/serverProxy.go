package proxy

import (
	"io"
	"log"
	"net"
	"strconv"
)

func Listen() {
	//addr := net.TCPAddr{
	//	IP:   net.ParseIP("127.0.0.1"),
	//	Port: 23333,
	//}
	//serv, err := net.ListenTCP("tcp", &addr)
	serv, err := net.Listen("tcp",
		net.JoinHostPort("127.0.0.1", "23333"))
	if err != nil {
		log.Println("server listen error")
		return
	}
	log.Println("start accept")
	for {
		conn, err := serv.Accept()
		log.Println("accepted a connection!")
		if err != nil {
			log.Println("server accept error")
			return
		}
		go hello(conn)
	}
}

func hello(lconn net.Conn) {
	b := make([]byte, 4096)
	_, err := lconn.Read(b)
	if err != nil {
		log.Println(err.Error())
		return
	}
	if b[0] == 0xDD {
		if b[1] != 0x00 {
			log.Println("ack 0x00 error")
			return
		}
		switch b[2] {
		case 0x01:
			log.Println("write response")
			lconn.Write([]byte{0xDD, 0x00, 0x01})

		}
	} else {
		log.Println("error")
	}

	n, err := lconn.Read(b)

	if b[0] == 0xDD {
		if b[1] == 0x01 {
			host := string(b[2 : n-2])
			port := strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))
			///todo bug
			log.Println(host, ":", port, "==================")
			tcpProxy(lconn, host, port)
		} else {
			log.Println("ack 0x01 error")
			return
		}
	}
}

//test function
func tcpProxy(lconn net.Conn, host, port string) {
	///todo 加密解密数据
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
	//n, err = lconn.Read(b)
	//println(string(b))
	//rconn.Write(b)

	//复制left请求到right
	copyReqRes := func(des, src net.Conn) {
		_, err := io.Copy(des, src)
		if err != nil {
			log.Println("error : ", err.Error())
		}
	}

	go copyReqRes(rconn, lconn)
	copyReqRes(lconn, rconn)

	log.Println("req && res copy over")
}

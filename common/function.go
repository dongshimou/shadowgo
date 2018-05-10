package common

import (
	"io"
	"log"
	"net"
)

//复制left请求到right
func CopyData(dst, src net.Conn) {
	/*
		_, err := io.Copy(des, src)
		if err != nil {
			log.Println("error : ", err.Error())
		}
		log.Println("go copy over!!!!!!")
	*/
	buf := make([]byte, 4096)
	var written int64 = 0
	var err error

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	log.Print(err)
}

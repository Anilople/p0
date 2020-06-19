// Implementation of a MultiEchoServer. Students should write their code in this file.

package p0

import (
	"net"
	"strconv"
)

type multiEchoServer struct {
	// TODO: implement this!
	es *EchoServer
}

// New creates and returns (but does not start) a new MultiEchoServer.
func New() MultiEchoServer {
	// TODO: implement this!
	mes := &multiEchoServer{
		nil,
	}
	return mes
}

// 启动一个服务，暴露出端口
// 会被调用多次
func (mes *multiEchoServer) Start(port int) error {
	// TODO: implement this!
	// 发消息给客户端
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	} else {
		mes.es = NewEchoServer(listener)
	}
	return nil
}

// 关闭服务器
func (mes *multiEchoServer) Close() {
	// TODO: implement this!
	mes.es.Close()
}

// return: 当前连接的客户端数量
func (mes *multiEchoServer) Count() int {
	return mes.es.Count()
}

// TODO: add additional methods/functions below!
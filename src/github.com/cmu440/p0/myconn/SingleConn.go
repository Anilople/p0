// 处理单个连接
// cd 到此文件所在目录下
// go install
package myconn

import (
	"bufio"
	"net"
)

// 对单个连接的抽象
type SingleConn struct {
	conn net.Conn
	// 通知服务器连接关闭
	connExit chan<- net.Conn
}

func New(
	conn net.Conn,
	receiveLineChan chan<- string,
	connChan chan<- net.Conn,
	) *SingleConn {
	sc := &SingleConn{
		conn:     conn,
		connExit: connChan,
	}
	go sc.handleReadFromClient(receiveLineChan)
	return sc
}

// 接收客户端发来的多行消息
func (sc *SingleConn) handleReadFromClient(receiveLineChan chan<- string) {
	defer func() {
		sc.connExit <-sc.conn
	}()

	reader := bufio.NewReader(sc.conn)
	for lineBytes, _, err := reader.ReadLine();
		err == nil;
	lineBytes, _, err = reader.ReadLine() {
		// 提供接收到的消息
		receiveLineChan <- string(lineBytes)
	}

}

// send 发送1行消息到客户端(不带换行符)
func (sc *SingleConn) Send(line string) {
	sc.conn.Write([]byte(line + "\n"))
}

// Close 关闭和客户端的连接
func (sc *SingleConn) Close() {
	sc.conn.Close()
}



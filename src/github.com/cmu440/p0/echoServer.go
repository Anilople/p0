// 单个 echo server Handler
// 注意使用chan []byte会发生消息乱序（byte为单位的交叉），原因未知
// 但是使用chan string就不会发生这个问题

// 并发写入map会发生错误fatal error: concurrent map writes
package p0

import (
	"fmt"
	"net"

	// 由于未知原因，使用 "./myconn" 不生效
	"github.com/cmu440/p0/myconn"
)

type EchoServer struct {
	listener net.Listener
	// 接受连接进入的消息
	connsEnter chan net.Conn
	// 接受连接断开的消息
	connsExit chan net.Conn
	// 连接池
	conns map[net.Conn]*myconn.SingleConn
	// 服务关闭的消息
	closeChan chan bool
	// 接受到来自客户端的数据，按 行 为单位
	receiveLineChan chan string
}

// 不断读取输入，然后广播
func (es *EchoServer) broadcastLoop() {
	for line, ok := <-es.receiveLineChan; ok; line, ok = <-es.receiveLineChan {
		go es.broadcast(line)
	}
}

// 将1条消息（无换行符，需手动添加），广播给所有客户端
func (es *EchoServer) broadcast(line string) {
	for _, singleConn := range es.conns {
		// 只管送消息，后续怎么处理，由更下一层的实现负责
		singleConn.Send(line)
	}
}

// 处理连接池
func (es *EchoServer) handleConnPool() {
	isOpen := true
	for isOpen {
		select {
		case conn := <- es.connsEnter:
			// 新连接进入
			sc := myconn.New(conn, es.receiveLineChan, es.connsExit)
			es.conns[conn] = sc
		case conn := <- es.connsExit:
			// 连接断开
			es.conns[conn].Close()
			// 从连接池中删除
			delete(es.conns, conn)
		case <- es.closeChan:
			isOpen = false
			break
		}
	}
	// 处理后事
	connsBackup := es.conns
	es.conns = nil
	for _, sc := range connsBackup {
		sc.Close()
	}
}

func (es *EchoServer) handleListener() {
	conn, err := es.listener.Accept()
	for err == nil {
		// 新连接进入
		es.connsEnter <- conn
		conn, err = es.listener.Accept()
	}
	// 已经close
	fmt.Println("listener", es.listener, "meet some error:", err)
}

func NewEchoServer(listener net.Listener) *EchoServer {
	es := &EchoServer{
		listener,
		make(chan net.Conn),
		make(chan net.Conn),
		make(map[net.Conn]*myconn.SingleConn),
		make(chan bool),
		make(chan string),
	}
	// 处理连接池
	go es.handleConnPool()
	// 处理广播
	go es.broadcastLoop()
	// 处理listener
	go es.handleListener()
	return es
}

// 关闭服务器
func (es *EchoServer) Close() {
	// TODO: implement this!
	// 发送关闭的消息
	es.closeChan <- true
}

// return: 当前连接的客户端数量
func (es *EchoServer) Count() int {
	return len(es.conns)
}

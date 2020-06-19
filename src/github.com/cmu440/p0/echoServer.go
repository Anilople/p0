// 单个 echo server Handler
// 注意使用chan []byte会发生消息乱序（byte为单位的交叉），原因未知
// 但是使用chan string就不会发生这个问题

// 并发写入map会发生错误fatal error: concurrent map writes
package p0

import (
	"bufio"
	"fmt"
	"net"
)

// 锁
var mutexP *Mutex = NewMutex()

type EchoServer struct {
	listener net.Listener
	isClose  bool
	// 还在保持可读连接的客户端
	readerConns map[net.Conn]bool
	// 还在保持可写连接的客户端
	writerConns map[net.Conn]bool
	// 接受到来自客户端的数据，按 行 为单位
	receiveLineChan chan string
}

// 不断读取输入，然后广播
func (es *EchoServer) broadcastLoop() {
	for {
		line := <-es.receiveLineChan
		es.broadcast(line)
	}
}

// 将1条消息（无换行符，需手动添加），广播给所有客户端
func (es *EchoServer) broadcast(line string) {
	for conn := range es.writerConns {
		_, err := conn.Write([]byte(line + "\n"))
		if err != nil {
			// 出现问题
			func() {
				mutexP.Lock()
				defer mutexP.UnLock()
				delete(es.writerConns, conn)
			}()
		}
	}
}

func (es *EchoServer) handleListener() {
	conn, err := es.listener.Accept()
	for err == nil {
		go es.handleReaderConn(conn)
		// 添加写入的conn
		func() {
			mutexP.Lock()
			defer mutexP.UnLock()
			es.writerConns[conn] = true
		}()
		conn, err = es.listener.Accept()
	}
	// 已经close
	fmt.Println("listener", es.listener, "meet some error:", err)
}

func (es *EchoServer) handleReaderConn(conn net.Conn)  {
	// 防止崩溃无法释放锁
	func() {
		mutexP.Lock()
		defer mutexP.UnLock()
		es.readerConns[conn] = true
	}()

	defer func() {
		mutexP.Lock()
		defer mutexP.UnLock()
		delete(es.readerConns, conn)
	}()
	//defer conn.Close()
	reader := bufio.NewReader(conn)
	lineBytes, _, err := reader.ReadLine()
	for err == nil {
		// 提供接收到的消息
		es.receiveLineChan <- string(lineBytes)

		// 继续接收
		lineBytes, _, err = reader.ReadLine()
	}
	//fmt.Println("read end", conn, err)
}

func NewEchoServer(listener net.Listener) *EchoServer {
	es := &EchoServer{
		listener,
		false,
		make(map[net.Conn]bool),
		make(map[net.Conn]bool),
		make(chan string, 100),
	}
	// 处理广播
	go es.broadcastLoop()
	// 处理listener
	go es.handleListener()
	return es
}

// 关闭服务器
func (es *EchoServer) Close() {
	// TODO: implement this!
	es.listener.Close()
	for conn := range es.readerConns {
		conn.Close()
	}
	es.readerConns = nil
	for conn := range es.writerConns {
		conn.Close()
	}
	es.isClose = true
	es.writerConns = nil
}

// return: 当前连接的客户端数量
func (es *EchoServer) Count() int {
	readerCount := len(es.readerConns)
	writerCount := len(es.writerConns)
	fmt.Println(readerCount, writerCount)
	if readerCount < writerCount {
		return readerCount
	} else {
		return writerCount
	}
}

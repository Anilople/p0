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
var mutexP = NewMutex()

type EchoServer struct {
	listener net.Listener
	isClose  bool
	// 还在保持可读连接的客户端
	readerConns map[net.Conn]bool
	// 还在保持可写连接的客户端
	writerConns map[net.Conn]chan string
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
	for _, lineChan := range es.writerConns {
		lineChan <- line
	}
}

func (es *EchoServer) handleListener() {
	conn, err := es.listener.Accept()
	for err == nil {
		// 防止崩溃无法释放锁
		func() {
			mutexP.Lock()
			defer mutexP.UnLock()
			es.readerConns[conn] = true
			// 添加写入的conn
			es.writerConns[conn] = make(chan string, 100)
		}()
		go es.handleReaderConn(conn)
		go es.handleWriterConn(conn, es.writerConns[conn])
		conn, err = es.listener.Accept()
	}
	// 已经close
	fmt.Println("listener", es.listener, "meet some error:", err)
}

func (es *EchoServer) handleReaderConn(conn net.Conn) {
	defer conn.Close()
	defer func() {
		mutexP.Lock()
		defer mutexP.UnLock()
		delete(es.readerConns, conn)
	}()
	reader := bufio.NewReader(conn)
	for lineBytes, _, err := reader.ReadLine();
		err == nil;
	lineBytes, _, err = reader.ReadLine() {
		// 提供接收到的消息
		es.receiveLineChan <- string(lineBytes)
	}
}

func (es *EchoServer) handleWriterConn(conn net.Conn, lineChan <-chan string) {
	defer conn.Close()
	defer func() {
		mutexP.Lock()
		defer mutexP.UnLock()
		delete(es.writerConns, conn)
	}()
	// 从channel中读消息，发送
	for line, ok := <-lineChan; !es.isClose && ok; line, ok = <-lineChan {
		_, err := conn.Write([]byte(line + "\n"))
		if err != nil {
			// 出现问题
			break
		}
	}
}

func NewEchoServer(listener net.Listener) *EchoServer {
	es := &EchoServer{
		listener,
		false,
		make(map[net.Conn]bool),
		make(map[net.Conn]chan string),
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

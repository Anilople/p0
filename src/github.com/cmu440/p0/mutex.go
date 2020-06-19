package p0

// 用channel来实现锁
type Mutex struct {
	boolChan chan bool
}

func NewMutex() *Mutex {
	m := &Mutex{make(chan bool, 1)}
	// m.Lock()
	m.UnLock()
	return m
}

// 加锁
func (m *Mutex) Lock() {
	// m.boolChan <- true
	<-m.boolChan
}

// 解锁
func (m *Mutex) UnLock() {
	// <-m.boolChan
	m.boolChan <- true
}

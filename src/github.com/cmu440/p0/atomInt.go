// 并发安全的计数器
package p0

type AtomInt struct {
	count int
	mutex *Mutex
}

// 构造
func NewAtomInt() *AtomInt {
	return &AtomInt{0, NewMutex()}
}

// +1
func (ai *AtomInt) inc()  {
	ai.mutex.Lock()
	ai.count++
	ai.mutex.UnLock()
}

// -1
func (ai *AtomInt) dec() {
	ai.mutex.Lock()
	ai.count--
	ai.mutex.UnLock()
}

func (ai *AtomInt) reset()  {
	ai.mutex.Lock()
	ai.count = 0
	ai.mutex.UnLock()
}

func (ai *AtomInt) get() int {
	count := -1
	ai.mutex.Lock()
	count = ai.count
	ai.mutex.UnLock()
	return count
}
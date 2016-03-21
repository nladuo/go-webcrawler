package model

type ThreadManager struct {
	threadManagerChan chan byte
	tag               int
	thread_num        int
}

func NewThreadManager(thread_num int) *ThreadManager {
	if thread_num == 0 { //cannot set the channel to unbufferred channel
		thread_num = 1
	}
	var tm ThreadManager
	tm.threadManagerChan = make(chan byte, thread_num)
	tm.thread_num = thread_num
	tm.tag = 0

	return &tm
}

//get the occupation of a downloader
func (this *ThreadManager) GetOccupation() int {
	this.threadManagerChan <- byte(0)
	return this.getTag()
}

func (this *ThreadManager) FreeOccupation() {
	<-this.threadManagerChan
}

func (this *ThreadManager) getTag() int {
	this.tag++
	if this.tag > this.thread_num {
		this.tag = 1
	}
	return this.tag
}

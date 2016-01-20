package scheduler

func cleanChan(ch chan byte) {
	for i := 0; i < len(ch); i++ {
		<-ch
	}
}

package chantable

import (
	"fmt"
	"time"
)

var (
	RecoverErrorChan = make(chan error, 1)
)

func Go(f func() error) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			RecoverErrorChan <- fmt.Errorf("%s\n", err)
		}
	}()
	//循环
	if fnErr := f(); fnErr != nil {
		select {
		case RecoverErrorChan <- fnErr:
		case <-time.After(time.Millisecond * 50):
			return
		}
	}
}

/*
@Time : 2021/4/28 下午3:32
@Author : MuYiMing
@File : mode_test
@Software: GoLand
*/
package test

import (
	"fmt"
	"github.com/MIM1993/chantable"
	"log"
	"testing"
	"time"
)

type Sub1 struct {
	SubName string
}

func (s *Sub1) Quit() {
	panic("implement me")
}

func (s *Sub1) Out() error {
	panic("implement me")
}

func (s *Sub1) OnMessage(msg *chantable.Message) {
	fmt.Printf("[Subscriber:%s ,Topic:%s]:%v \n", s.SubName, msg.Topic, msg.Payload)
}

func (s *Sub1) Name() string {
	return s.SubName
}

func TestCommunication(t *testing.T) {
	topicName := "mim"
	msghdl := chantable.NewMsgHandler(1)
	err := msghdl.RegisteredTopic(topicName)
	if err != nil {
		log.Panicf("RegisteredTopic err:%v", err)
	}

	err = msghdl.Start()
	if err != nil {
		log.Panicf("Start err:%v", err)
	}

	time.Sleep(time.Second * 1)

	t.Run("sub1", func(t *testing.T) {
		s1 := &Sub1{SubName: "mim"}
		err = msghdl.AddSubscriber(topicName, s1)
		if err != nil {
			log.Panicf("AddSubscriber err:%v", err)
		}

		for i := 0; i < 10; i++ {
			err = msghdl.PublishOrder(topicName, fmt.Sprintf("hello world, number:%d", i))
			if err != nil {
				log.Panicf("PublishOrder err:%v", err)
			}
		}
	})

	time.Sleep(100 * time.Second)

}

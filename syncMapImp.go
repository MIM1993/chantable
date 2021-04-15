package chantable

import (
	"fmt"
	"sync"
)

//消息表实例
type msgHandler struct {
	topicNum int32
	//每个topic可以添加订阅者的上线

	chanNum int32
	//每个chan可以缓存的容量，为0则使用默认值

	l *sync.RWMutex
	//操作共享数据时枷锁

	topicTab *sync.Map
	/*
		map[string][]Subyscriber
		是一个线程安全的map，用于维持topic和订阅者这个topic的所有订阅者的联系
	*/

	//once  *sync.Once 待定

	chanTab *sync.Map
	//存储每个topic对应的channel

	orderChan chan *Message
	//顺序管道  Message是一个struct，保存了数据需要发送到的topic和具体数据

	quitC chan struct{}
	//退出监听chan

	status uint8
	/*表示当前消息表的状态
	0   close
	1   running
	*/
}

var _ MessageTaber = &msgHandler{}

//创建一个消息表
func NewMsgHandler(topicNum int32) MessageTaber {
	return &msgHandler{
		topicNum:  topicNum,
		l:         &sync.RWMutex{},
		topicTab:  &sync.Map{},
		chanTab:   &sync.Map{},
		orderChan: make(chan *Message, 1),
		quitC:     make(chan struct{}),
		status:    0,
	}
}

func (msg *msgHandler) WithChanNum(chanNum int32) {
	msg.chanNum = chanNum
}

func (msg *msgHandler) Start() error {
	//启动守护goroutine用于顺序数据发送
	fn := func() error {
		for {
			select {
			case _ = <-msg.orderChan:
				//将消息发送到用户
			case <-msg.quitC:
				//回收资源，关闭消息表
				close(msg.orderChan)
				return nil
			}
		}
	}
	//开启goroutine
	go Go(fn)
	msg.status = 1
	return nil
}

func (msg *msgHandler) RegisteredTopic(topic string) error {
	//加锁，解锁
	msg.l.Lock()
	defer msg.l.Unlock()

	//尝试从表中获取topic
	subs, ok := msg.topicTab.Load(topic)
	if subs == nil || !ok {
		//create topic and save topicMap
		msg.topicTab.Store(topic, make([]Subscriber, 0))
		//create chanMap and save chan
		var num int32
		if msg.chanNum != 0 {
			num = msg.chanNum
		} else {
			num = DEFULTCHANNUM
		}
		msgC := make(chan *Message, num)
		msg.chanTab.Store(topic, msgC)
		fn := func() error {
			for {
				select {
				case m, ok := <-msgC:
					if !ok {
						msg.l.Lock()
						//删除topicTab中topic
						msg.topicTab.Delete(topic)
						msg.l.Unlock()
						return nil
					}
					//将消息发送到用户
					_ = m
				case <-msg.quitC:
					//回收资源，关闭消息表
					close(msgC)
					return nil
				}
			}
		}
		//start watch chan goroutine
		go Go(fn)
	} else {
		//topic exist
		return fmt.Errorf("topic:%s exist,cannot be added repeatedly", topic)
	}
	return nil
}

func (msg *msgHandler) DelTopic(topic string) error {
	//加锁，解锁
	msg.l.Lock()
	defer msg.l.Unlock()

	//尝试从表中获取topic
	subs, ok := msg.topicTab.Load(topic)
	if subs == nil || !ok {
		return nil
	} else {
		//close chan
		c, ok := msg.chanTab.Load(topic)
		if c != nil && ok {
			msgC, ok := c.(chan *Message)
			if !ok {
				return fmt.Errorf("type is not chan *Message")
			}
			close(msgC)
			msg.chanTab.Delete(topic)
		}
	}
	return nil
}

//todo: error
func (msg *msgHandler) AllTopic() ([]string, error) {
	msg.l.RLock()
	defer msg.l.RUnlock()

	var topicArr []string
	msg.topicTab.Range(func(key, _ interface{}) bool {
		tmpStr, ok := key.(string)
		if ok {
			topicArr = append(topicArr, tmpStr)
			return true
		}
		return false
	})
	return topicArr, nil
}

func (msg *msgHandler) AddSubscriber(topic string, sub ...Subscriber) error {
	//加锁，解锁
	msg.l.Lock()
	defer msg.l.Unlock()

	//尝试从表中获取topic
	subs, ok := msg.topicTab.Load(topic)
	if subs != nil && ok {
		arr, ok := subs.([]Subscriber)
		if !ok {
			return fmt.Errorf("type error")
		}
		//判断是否有Subyscriber重复
		for _, v := range sub {
			if isRedundant(arr, v) {
				continue
			}
			arr = append(arr, v)
		}
	}
	return fmt.Errorf("topic %s not exit or other cause", topic)
}

//判断Subscriber是否重复
func isRedundant(subs []Subscriber, sub Subscriber) bool {
	for _, s := range subs {
		if s == sub {
			return true
		}
	}
	return false
}

func (msg *msgHandler) DelSubscriber(topic string, sub Subscriber) error {
	//加锁，解锁
	msg.l.Lock()
	defer msg.l.Unlock()

	subs, ok := msg.topicTab.Load(topic)
	if ok && subs != nil {
		arr, ok := subs.([]Subscriber)
		if !ok {
			return fmt.Errorf("type error")
		}
		for i, s := range arr {
			if s == sub {
				copy(arr[i:], arr[i+1:])
				arr = arr[:len(arr)-1]
				msg.topicTab.Store(topic, arr)
				return nil
			}
		}
	}
	return fmt.Errorf("topic %s not exit", topic)
}

func (msg *msgHandler) StatisticalSubscriber(topic string) (int, error) {
	msg.l.RLock()
	defer msg.l.RUnlock()

	subs, ok := msg.topicTab.Load(topic)
	if ok && subs != nil {
		arr, ok := subs.([]Subscriber)
		if !ok {
			return -1, fmt.Errorf("type error")
		}
		return len(arr), nil
	}
	return -1, fmt.Errorf("topic %s not exit", topic)
}

func (msg *msgHandler) Publish(topic string, Payload interface{}) error {
	msg.l.RLock()
	defer msg.l.RUnlock()

	if msg.status == 0 {
		return fmt.Errorf("msg table is closed")
	}

	msgC, ok := msg.chanTab.Load(topic)
	if msgC == nil || !ok {
		return fmt.Errorf("get topic error,maybe topic not exist")
	}

	msgChan, ok := msgC.(chan *Message)

	msgChan <- &Message{
		Topic:   topic,
		Payload: Payload,
	}
	return nil
}

func (msg *msgHandler) PublishOrder(topic string, Payload interface{}) error {
	msg.l.RLock()
	defer msg.l.RUnlock()

	if msg.status == 0 {
		return fmt.Errorf("msg table is closed")
	}

	//放入顺序管道
	msg.orderChan <- &Message{
		Topic:   topic,
		Payload: Payload,
	}
	return nil
}

//从容中出
func (msg *msgHandler) Close() error {
	//关闭总线,不接收消息
	msg.status = 0

	//关闭退出信息chan
	//close(msg.quitC)

	//循环关闭所有监听管道
	msg.chanTab.Range(func(_, value interface{}) bool {
		c, ok := value.(chan *Message)
		if ok {
			close(c)
			return true
		}
		return false
	})

	//关闭顺序chan
	close(msg.orderChan)

	return nil
}

//强制中出
func (msg *msgHandler) Kill() {
	close(msg.quitC)
}

package chantable

//消息表支持方法
type MessageTaber interface {
	// RegisteredTopic 注册topic
	RegisteredTopic(string) error
	// DelTopic 删除topic
	DelTopic(string) error
	// AllTopic 查询所有topic
	AllTopic() ([]string, error)

	// AddSubscriber 添加topic下订阅者
	AddSubscriber(string,...Subscriber) error
	// DelSubscriber 删除topic下订阅者
	DelSubscriber(string,Subscriber) error

	// StatisticalSubscriber 查询topic订阅人数
	StatisticalSubscriber(string) (int, error)

	// Publish 并发发送消息
	Publish(string, interface{}) error
	// PublishOrder 顺序发送消息
	PublishOrder(string, interface{}) error

	// Start 启动消息表
	Start()error
	// Close 关闭总消息表
	Close() error

	//kill Force Close
	Kill()
}

//用户（订阅者）支持方法
type Subscriber interface {
	//订阅的topic有消息时执行
	OnMessage(*Message)
	//当订阅者被删除时调用
	Quit()
	//订阅者主动取消链接
	Out() error
}

//具体消息数据结构
type Message struct {
	Topic   string
	Payload interface{}
}

const (
	DEFULTCHANNUM = 10240 //暂定
)
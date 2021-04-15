# chantable
功能：消息总线，订阅，发送

#### 接口定义：

```go
//消息表支持方法
type MessageTab interface{
    //注册topic
    RegisteredTopic(string)error
    //删除topic
    DelTopic(string)error
    //查询所有topic
    AllTopic()([]string,error)
    
    //添加topic下订阅者
    AddSubscriber(Subyscriber)error
    //删除topic下订阅者
    DelSubscriber(Subyscriber)error
    //修改topic订阅人数
    //ModifySubscriber()(int,error)
    //查询topic订阅人数
    StatisticalSubscriber()(int,error)
    
    //并发发送消息
    Publish(string,interface{})error
    //顺序发送消息
    PublishOrder(string,interface{})error
    
    //关闭总消息表
    Close()error
    
    //强制关闭
    Kill()
}
```



```go
//用户（订阅者）支持方法
type Subyscriber interface{
    //订阅的topic有消息时执行
    OnMessage(*Message)
    //当订阅者被删除时调用
    Quit()
    //订阅者主动取消链接
    Out()error    
}
```



#### 数据结构定义：

```go
//消息表实例
type msgHandler struct{
    num          int32 
    //每个topic可以添加订阅者的上线
   
    l  		     *sync.RWMutex
    //操作共享数据时枷锁
    
    topicTab     *sync.Map
    /*
    	map[string][]Subyscriber
    	是一个线程安全的map，用于维持topic和订阅者这个topic的所有订阅者的联系
    */
    
    //once  *sync.Once 待定
    
    chanTab      *sync.Map
    //存储每个topic对应的channel
    
    orderChan    chan *Message 
    //顺序管道  Message是一个struct，保存了数据需要发送到的topic和具体数据
    
    quitC        chan struct{}
    //退出监听chan
    
	status       uint8
    /*表示当前消息表的状态
    	0   close
    	1   running
    */
}
```



```go
//具体消息数据结构
type Message struct {
	Topic   string
	Payload interface{}
}
```










package eventbus

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Subscriber interface {
	Name() string
	Execute()
	Result() string
	Mode() string
}

type EventBus struct {
	locker sync.RWMutex

	// 缓存EVENT
	// PENDING队列是一个优先级队列
	pending *Heap
	posting bool
	active  chan struct{}

	// 正常执行的事件，FIFO队列
	// 初始化时，会从数据库中恢复正在执行的事件
	// 正在执行的事件数量不能超出设定的阈值（但是事件属性为JustDoIt的事件不受这个约束）
	// 会定时巡检队列中事件的状态，如果事件执行成功、失败或者超时，直接出队
	running *list.List

	// 订阅者
	// KEY是EVENT的Type+Name，VALUE是所有订阅该事件的订阅者
	// 当事件发生时，会通知所有的订阅者消费事件
	// 事件的订阅者由配置文件配置
	topices     map[string]([]Subscriber)
	subscribers map[string]int
}

func NewEventBus() *EventBus {
	return &EventBus{
		pending: NewHeap(func(x interface{}) string {
			e := x.(*Event)
			return e.id
		}),
		running:     list.New(),
		active:      make(chan struct{}),
		topices:     make(map[string][]Subscriber),
		subscribers: make(map[string]int),
	}
}

func (bus *EventBus) Subscribe(topic string, subscriber Subscriber) (err error) {
	if _, ok := bus.subscribers[subscriber.Name()]; ok {
		return errors.New("found subscriber, " + subscriber.Name())
	}

	bus.locker.Lock()
	defer bus.locker.Unlock()

	bus.topices[topic] = append(bus.topices[topic], subscriber)
	return
}

func (bus *EventBus) Post(event *Event) {
	fmt.Println("add event to pending queue")
	bus.pending.Add(event, event.priority)

	bus.locker.RLock()
	defer bus.locker.RUnlock()

	if !bus.posting {
		fmt.Println("active post event")
		bus.active <- struct{}{}
	}
}

func (bus *EventBus) Run() {
	for {
		select {
		case <-bus.active:
		case <-time.After(time.Second * 1):
			fmt.Println("timeout")
			if bus.running.Len() != 0 {
				fmt.Println("change running to finished")
				e := bus.running.Front()
				ev := e.Value.(*Event)

				subscribers, ok := bus.topices[ev.topic]
				if !ok {
					break
				}

				for _, subscriber := range subscribers {
					subscriber.Result()
				}

				bus.running.Remove(e)
			}
		}

		fmt.Println("pop event from pending queue")
		bus.posting = true
		size := bus.pending.Size()
		for i := 0; i < size; i++ {
			e := bus.pending.Pop()
			ev := e.(*Event)

			bus.notify(ev)
		}

	}
}

func (bus *EventBus) notify(ev *Event) bool {
	// 不存在也直接消费掉
	subscribers := bus.topices[ev.topic]
	for _, subscriber := range subscribers {
		switch subscriber.Mode() {
		case "JustDoIt":
			subscriber.Execute()

		case "Background":
			// 限制后台执行的数量
			if bus.running.Len() > 5 {
				return false
			}

			subscriber.Execute()
			bus.running.PushBack(ev)
		}

	}

	return true
}

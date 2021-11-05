package eventbus

import (
	"fmt"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	bus := NewEventBus()
	bus.Subscribe("test", &subscriber{})

	go bus.Run()

	bus.Post(&Event{id: "1", priority: 0, topic: "test"})
}

type subscriber struct {
	name        string
	runningTime int
	mode        string
}

func NewSubscriber(name, mode string, running int) *subscriber {
	return &subscriber{
		name:        name,
		runningTime: running,
		mode:        mode,
	}
}

func (s *subscriber) Name() string {
	return s.name
}

func (s *subscriber) Execute() {
	// 1~100s
	time.Sleep(time.Second * time.Duration(s.runningTime))
}

func (s *subscriber) Result() string {
	// 1 ~ 100s
	return ""
}

func (s *subscriber) Mode() string {
	return s.mode
}

func Test1(t *testing.T) {
	// 1000 event, 1 subscriber

}

func BenchmarkTest2(b *testing.B) {
	bus := NewEventBus()
	_ = bus.Subscribe("test", NewSubscriber("test", "", 1))

	go bus.Run()

	t1 := time.Now()

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("%d", i)
		bus.Post(&Event{id: id, priority: 1, topic: "test"})

	}

	for {
		if bus.running.Len() != 0 {
			continue
		}

		break
	}

	fmt.Printf("run:%d ms\n", time.Now().Sub(t1).Milliseconds())
}

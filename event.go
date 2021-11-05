package eventbus

type Event struct {
	priority int
	topic    string
	id       string
}

package log

type chainData struct {
	messages []Message
}

func Chain(msg Message) *chainData {
	return &chainData{
		messages: []Message{msg},
	}
}

func (chain *chainData) Add(msg Message) {
	chain.messages = append(chain.messages, msg)
}

func (chain *chainData) Write() {
	LOG(chain.messages...)
	chain.messages = nil
}

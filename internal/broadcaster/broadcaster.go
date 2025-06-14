package broadcaster

import "fmt"

// based on tutorial from https://www.youtube.com/watch?v=MuyYQWeBTyU
type Broadcaster struct {
	// holds broadcast messages
	Broadcast chan string
	// accept new connections
	NewConnection chan chan string
	CloseConnection chan chan string
	Connections map[chan string]int32
}
func (b *Broadcaster) Listen() {
	var seq int32 = 0
	for {
		select {
		// broadcast messages to all connections
		case message := <-b.Broadcast:
			for connection := range b.Connections {
				connection <- message
			}
		// new connection
		case connection := <-b.NewConnection:
			b.Connections[connection] = seq
			seq++
			fmt.Printf("New connection %v\n", connection)
		// close connections
		case connection := <-b.CloseConnection:
			delete(b.Connections, connection)
			fmt.Printf("Closed connection %v\n", connection)
		}
	}
}
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		Broadcast: make(chan string),
		NewConnection: make(chan chan string),
		CloseConnection: make(chan chan string),
		Connections: make(map[chan string]int32),
	}

}

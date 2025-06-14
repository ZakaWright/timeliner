package broadcaster

//import "fmt"

// based on tutorial from https://www.youtube.com/watch?v=MuyYQWeBTyU
type Broadcaster struct {
	/*
	// holds broadcast messages
	Broadcast chan string
	// accept new connections
	NewConnection chan chan string
	CloseConnection chan chan string
	Connections map[chan string]int32
	*/

	Broadcaster chan Message
	RegisterClient chan Client
	UnregisterClient chan Client
	Clients map[int64]map[chan string] struct{}
	
}

type Client struct {
	IncidentID int64
	Channel chan string
}

type Message struct {
	IncidentID int64 
	Message string
}

func (b *Broadcaster) Listen() {
	//var seq int32 = 0
	for {
		select {
		// broadcast messages to all connections
		/*
		case message := <-b.Broadcast:
			for connection := range b.Connections {
				connection <- message
			}
		// new connection
		case connection := <-b.NewConnection:
			b.Connections[connection] = seq
			seq++
		// close connections
		case connection := <-b.CloseConnection:
			delete(b.Connections, connection)
		}
		*/
		case client := <-b.RegisterClient:
			if b.Clients[client.IncidentID] == nil {
				b.Clients[client.IncidentID] = make(map[chan string]struct{})
			}
			b.Clients[client.IncidentID][client.Channel] = struct{}{}
		case client := <-b.UnregisterClient:
			if chans, ok := b.Clients[client.IncidentID]; ok {
				delete(chans, client.Channel)
				close(client.Channel)
				if len(chans) == 0 {
					delete(b.Clients, client.IncidentID)
				}
			}
		case message := <- b.Broadcaster:
			for channel := range b.Clients[message.IncidentID] {
				select {
				case channel <- message.Message:
				default:
					// skip if client isn't ready
				}
			}
		}
	}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		/*
		Broadcast: make(chan string),
		NewConnection: make(chan chan string),
		CloseConnection: make(chan chan string),
		Connections: make(map[chan string]int32),
		*/
		Broadcaster: make(chan Message),
		RegisterClient: make(chan Client),
		UnregisterClient: make(chan Client),
		Clients: make(map[int64]map[chan string]struct{}),
		}

}

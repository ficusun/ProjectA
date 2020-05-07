package wsserver

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

// Connections - []Client
type Connections []Client

// id for user
var lastID int

// Add - func for add new client for Clients array
// expected WS conn
func (c *Connections) Add(conn *websocket.Conn) {

	var Client Client = Client{
		ID:   lastID,
		Send: make(chan []byte, maxMessageSize),
	}

	Client.Conn = conn
	Client.Status = true
	*c = append(*c, Client)
	lastID++
	c.PushOnlineClientsToChat()
}

// GetClients - func GetClients() Clients
// return array of *Clients
func (c *Connections) GetClients() Connections {
	return *c
}

// GetClientsID - func GetClients()
// return []int
func (c *Connections) GetClientsID(ID int) []int {

	client := make([]int, 0, len(*c))

	if len(*c) > 0 {
		for _, cl := range *c {
			if cl.Status == true {
				client = append(client, cl.ID)
			}
		}
	}
	return client // strings.Join(client, " ")
}

// DelByID - delete client from Clients array by id
// expected id (int)
func (c *Connections) DelByID(id int) {
	switch id {
	case 0:
		*c = append((*c)[1:])
	case len(*c):
		*c = append((*c)[0 : id-1])
	default:
		*c = append((*c)[:id], (*c)[id+1:]...)
	}
}

// CleanOffConn - remove client with off status
func (c *Connections) CleanOffConn() {

	tick := time.Tick(ClearPer * time.Millisecond)

	for range tick {
		if len(*c) > 0 {
			thisdel := make([]int, 0, MaxConnections)
			for i := range *c {
				if (*c)[i].Status == false {
					thisdel = append(thisdel, i)
				}
			}
			if len(thisdel) > 0 {
				for _, ID := range thisdel {
					// fmt.Println("before del: ", *c)
					(*c).DelByID(ID)
					// fmt.Println("after del: ", *c)
				}

				fmt.Println("remove bad connections: ", thisdel)
				thisdel = make([]int, 0, MaxConnections)
				c.PushOnlineClientsToChat()
			}
		}
	}
}

// GetOfflineClient - return ids offline client
func (c *Connections) GetOfflineClient() []int {
	TheseOff := make([]int, 0, MaxConnections)
	if len(*c) > 0 {
		for i := range *c {
			if (*c)[i].Status == false {
				TheseOff = append(TheseOff, (*c)[i].ID)
			}
		}
		return TheseOff
	}
	return TheseOff

}

type UsersOnline []UserOnline
type UserOnline struct {
	ID   int    `json:"id"`
	Nick string `json:"n"`
}

// PushOnlineClientsToChat - return ids offline client
func (c *Connections) PushOnlineClientsToChat() {
	TheseOn := make(UsersOnline, 0, MaxConnections)
	if len(*c) > 0 {
		for i := range *c {
			if (*c)[i].Status == true {
				TheseOn = append(TheseOn, UserOnline{(*c)[i].ID, (*c)[i].Nick})
			}
		}
		online, err := json.Marshal(TheseOn)
		if err != nil {
			fmt.Println("GetOnlineClients: ", err)
		}

		mes, err := json.Marshal(Letter{87654321, "2550", string(online)})
		if err != nil {
			log.Fatalln(err)
		}

		FromConnChan <- mes
	}
}

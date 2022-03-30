package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	*websocket.Conn
}

// WsJsonResponse defines the response sent back from the websocket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// WsJsonPayload defines the request we sent to the ws endpoint from the client
type WsJsonPayload struct {
	Action   string              `json:"action"`
	Message  string              `json:"message"`
	Username string              `json:"username"`
	Conn     WebSocketConnection `json:"-"`
}

var wsChan = make(chan WsJsonPayload)
var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "chat.html", nil)
	if err != nil {
		log.Println(err)
	}
}

//WsEndpoint upgrades http connection to a websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	logErr(err)

	var response WsJsonResponse
	response.Message = `<em><small>Connected to server!</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	logErr(err)

	go ListenForWs(&conn)
}

func ListenForWs(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsJsonPayload
	for {
		err := conn.ReadJSON(&payload)
		if err == nil {
			payload.Conn = *conn
			wsChan <- payload
		} else {
			fmt.Println("Error, closing connection", err)
			conn.Close()
		}
	}
}

func ListenToWsChannel() {
	var response WsJsonResponse

	for {
		e := <-wsChan

		switch e.Action {
		case "username":
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "list_current":
			fmt.Println("List current...")
			fmt.Println(e)
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)
		case "left":
			fmt.Println(e)
			response.Action = "list_users"
			delete(clients, e.Conn)
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)
		}
	}
}

func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}

	sort.Strings(userList)
	return userList
}

func broadcastToAll(response WsJsonResponse) {
	for client, name := range clients {
		fmt.Println("Handling " + name)
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket err")
			_ = client.Close()

			// Remove someone who disconnected from the web client
			delete(clients, client)
		} else {
			fmt.Println(response)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	logAndReturnErr(err)

	err = view.Execute(w, data, nil)
	logAndReturnErr(err)

	return nil
}

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func logAndReturnErr(err error) error {
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type Client struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

type CheckIn struct {
	Id     string `json:"id"`
	Status string `json:"status"`
}

var (
	pingInterval = 60 * time.Second
	upgrader     = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// keep track of clients connected through websocket
	clients    = make(map[*websocket.Conn]Client)
	clientsMux sync.Mutex

	// simulating a check-in table in a db
	checkIns    = make(map[string]CheckIn)
	checkInsMux sync.Mutex

	checkInId = uuid.New().String()
)

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// initialize check-in "db record"
	checkIns[checkInId] = CheckIn{
		Id:     checkInId,
		Status: "",
	}

	e.GET("/", func(c echo.Context) error {
		return c.File("index.html")
	})
	e.GET("/ws", handleWebSocket)
	e.GET("/clients", getClients)
	e.GET("/check-ins", getCheckIn)

	e.Logger.Fatal(e.Start(":8081"))
}

func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// add client to map
	clientsMux.Lock()
	clients[ws] = Client{
		Id:     uuid.New().String(),
		Status: "connected",
	}
	clientsMux.Unlock()
	c.Logger().Info("Client %s connected\n", clients[ws].Id)

	// gracefully and safely remove client from map when connection is closed
	defer func() {
		clientsMux.Lock()
		c.Logger().Info("Client %s disconnected\n", clients[ws].Id)
		delete(clients, ws)
		clientsMux.Unlock()
		ws.Close()
	}()

	// Channel to signal when the connection is closed
	done := make(chan struct{})

	// set read deadline
	if err := ws.SetReadDeadline(time.Now().Add(pingInterval)); err != nil {
		c.Logger().Error(err)
		return err
	}

	go func() {
		defer close(done)
		for {
			// Read
			messageType, msg, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.Logger().Error(err)
				}
				return
			}
			// reset read deadline
			if err := ws.SetReadDeadline(time.Now().Add(pingInterval)); err != nil {
				c.Logger().Error(err)
				return
			}

			// respond to ping with pong to keep the connection alive with client
			trimmedMsg := strings.TrimSpace(string(msg))
			if messageType == websocket.TextMessage && trimmedMsg == "ping" {
				// respond to ping with pong
				err = ws.WriteMessage(websocket.TextMessage, []byte("pong"))
				if err != nil {
					c.Logger().Error(err)
					return
				}
			} else {
				// parse JSON messsage to update the check-in status
				var message CheckIn
				if err := json.Unmarshal(msg, &message); err != nil {
					c.Logger().Errorf("Error parsing JSON: %v", err)
					continue
				}
				updateStatus(c, message)
			}
		}
	}()

	// Wait for the connection to be closed
	<-done
	return nil
}

func updateStatus(c echo.Context, message CheckIn) {
	c.Logger().Info("Updating status...")
	checkInsMux.Lock()
	defer checkInsMux.Unlock()

	checkIns[checkInId] = CheckIn{
		Id:     checkInId,
		Status: message.Status,
	}
	c.Logger().Info("Status updated")
}

// endpoint to get the check-in status
func getCheckIn(c echo.Context) error {
	checkInsMux.Lock()
	defer checkInsMux.Unlock()

	checkInList := make([]CheckIn, 0, len(checkIns))
	for _, checkIn := range checkIns {
		checkInList = append(checkInList, checkIn)
	}
	return c.JSON(http.StatusOK, checkInList)
}

// endpoint to get the list of connected clients through websocket
func getClients(c echo.Context) error {
	clientsMux.Lock()
	defer clientsMux.Unlock()

	clientsList := make([]Client, 0, len(clients))
	for _, client := range clients {
		clientsList = append(clientsList, client)
	}
	return c.JSON(http.StatusOK, clientsList)
}

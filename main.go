package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Message struct {
	Username string    `json:"username"`
	Text     string    `json:"text"`
	Time     time.Time `json:"time"`
}

var upgrader = websocket.Upgrader{}
var connections = make(map[*websocket.Conn]bool)
var messages = make(chan Message, 10)
var users = []User{{"alice", "1234"}, {"bob", "5678"}}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/login", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, u := range users {
			if user.Username == u.Username && user.Password == u.Password {
				c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
	})

	router.GET("/ws", func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("read error:", err)
				delete(connections, conn)
				break
			}
			messages <- msg
		}
	})

	go handleMessages()

	if err := router.Run(":3000"); err != nil {
		log.Fatal(err)
	}
}

func handleMessages() {
	for {
		msg := <-messages
		for conn := range connections {
			err := conn.WriteJSON(msg)
			if err != nil {
				log.Println("write error:", err)
				delete(connections, conn)
			}
		}
	}
}

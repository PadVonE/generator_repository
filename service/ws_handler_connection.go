package service

import (
	"generator/entity"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func (s *Service) WsHandleConnections(ctx *gin.Context) {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer ws.Close()

	s.WsClients[ws] = true

	for {
		var msg entity.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(s.WsClients, ws)
			break
		}
		// Здесь вы можете обрабатывать сообщения от клиента.
	}
}

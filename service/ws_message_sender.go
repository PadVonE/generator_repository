package service

import (
	"generator/entity"
	log "github.com/sirupsen/logrus"
)

func (s *Service) sendMessageToClients(msg entity.Message) {
	erroredClients := make([]*WsClient, 0)

	for wsClient := range s.WsClients {
		wsClient.Mutex.Lock()
		err := wsClient.Conn.WriteJSON(msg)
		wsClient.Mutex.Unlock()
		if err != nil {
			log.Printf("error: %v", err)
			wsClient.Conn.Close()
			erroredClients = append(erroredClients, wsClient)
		}
	}

	for _, wsClient := range erroredClients {
		delete(s.WsClients, wsClient)
	}
}

func (s *Service) LogError(text string) {
	s.sendMessageToClients(entity.Message{
		Text:  text,
		Color: "red",
		Name:  "[ERR]",
	})
}

func (s *Service) LogInfo(text string) {
	s.sendMessageToClients(entity.Message{
		Text:  text,
		Color: "blue",
		Name:  "[INFO]",
	})
}

func (s *Service) LogWarning(text string) {
	s.sendMessageToClients(entity.Message{
		Text:  text,
		Color: "yellow",
		Name:  "[WARN]",
	})
}

func (s *Service) LogGlobal(text string) {
	s.sendMessageToClients(entity.Message{
		Text:  text,
		Color: "purple",
		Name:  "[GLOB]",
	})
}

func (s *Service) LogComplete(text string) {
	s.sendMessageToClients(entity.Message{
		Text:  text,
		Color: "green",
		Name:  "[GLOB]",
	})
}

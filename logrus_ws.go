package main

import (
	"generator/entity"
	"generator/service"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
	"sync"
)

type WsHook struct {
	Service *service.Service
	levels  []logrus.Level
	Mutex   sync.Mutex
}

func (hook *WsHook) Fire(entry *logrus.Entry) (err error) {
	message := entry.Message

	msg := entity.Message{
		Name:  "[" + strings.ToLower(entry.Level.String()) + "] ",
		Color: entry.Level.String(),
		Text:  message,
	}

	erroredClients := make([]*service.WsClient, 0)

	for wsClient := range hook.Service.WsClients {
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
		delete(hook.Service.WsClients, wsClient)
	}

	return err
}

func (hook *WsHook) Levels() []logrus.Level {
	return hook.levels
}

func initWebSocketHook(s *service.Service) {

	// Создание нового экземпляра WsHook
	hook := &WsHook{
		Service: s,
		levels:  logrus.AllLevels, // Перехватываем все уровни логов
	}

	// Добавление нового хука к Logrus
	logrus.AddHook(hook)
}

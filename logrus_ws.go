package main

import (
	"generator/entity"
	"generator/service"
	"github.com/sirupsen/logrus"
	"log"
	"strings"
)

type WsHook struct {
	Service *service.Service
	levels  []logrus.Level
}

func (hook *WsHook) Fire(entry *logrus.Entry) (err error) {
	message := entry.Message

	msg := entity.Message{
		Name:  "[" + strings.ToLower(entry.Level.String()) + "] ",
		Color: entry.Level.String(),
		Text:  message,
	}

	for client := range hook.Service.WsClients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(hook.Service.WsClients, client)
		}
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

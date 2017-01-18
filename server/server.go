package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	wdb "github.com/voidshard/wysteria/server/database"
	wsb "github.com/voidshard/wysteria/server/searchbase"
)

const (
	server_queue   = "WYSTERIA_SERVER_v01"
	ERR_UNKOWN_MSG = "Message not understood"
)

type WysteriaServer struct {
	GracefulShutdownTime time.Duration

	SettingsMiddleware wcm.MiddlewareSettings
	SettingsBackend    wdb.DatabaseSettings

	database   wdb.Database
	searchbase wsb.Searchbase
	middleware wcm.WysteriaMiddleware

	fromClients wcm.WysteriaSubscription
	fromServers wcm.WysteriaSubscription
}

func (s *WysteriaServer) Run() {
	log.Println("Running")
	for {
		select {
		case message := <-s.fromServers.Receive():
			log.Println(message)
		case message := <-s.fromClients.Receive():
			go s.handleClientMessage(&message)
		}
	}
}

func (s *WysteriaServer) handleClientMessage(message *wcm.Message) {
	var handler func([]byte) ([]byte, error)
	log.Println(message.Subject)

	switch message.Subject {

	// Search
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_COLLECTION:
		handler = s.handleFindCollection
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_ITEM:
		handler = s.handleFindItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_VERSION:
		handler = s.handleFindVersion
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_RESOURCE:
		handler = s.handleFindResource
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_LINK:
		handler = s.handleFindLink

	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_HIGHEST_VERSION:
		handler = s.handleFindHighestVersion

	// Create
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_COLLECTION:
		handler = s.handleCreateCollection
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_ITEM:
		handler = s.handleCreateItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_VERSION:
		handler = s.handleCreateVersion
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_RESOURCE:
		handler = s.handleCreateResource
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_LINK:
		handler = s.handleCreateLink

	// Update
	case s.SettingsMiddleware.RouteServer + wyc.MSG_UPDATE_ITEM:
		handler = s.handleUpdateItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_UPDATE_VERSION:
		handler = s.handleUpdateVersion

	// Delete
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_COLLECTION:
		handler = s.handleDelCollection
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_ITEM:
		handler = s.handleDelItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_VERSION:
		handler = s.handleDelVersion
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_RESOURCE:
		handler = s.handleDelResource

	// ?
	default:
		break
	}

	if handler == nil {
		s.sendError(message.Subject, message.Reply, errors.New(fmt.Sprintf("%s '%s'", ERR_UNKOWN_MSG, message.Subject)))
		return
	}

	reply, err := handler(message.Data)
	if err != nil {
		s.sendError(message.Subject, message.Reply, err)
	} else {
		s.middleware.Publish(message.Reply, reply)
	}
}

func (s *WysteriaServer) sendError(subject, reply string, err error) {
	log.Println(subject, err)
	s.middleware.Publish(reply, []byte(fmt.Sprintf("%s %s", wyc.WYSTERIA_SERVER_ERR, err.Error())))
}

func (s *WysteriaServer) Shutdown() {
	go s.close() // send a routine to kill off connections nicely

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	msg := "Shutdown request recieved, giving %s for connections to close gracefully"
	log.Println(fmt.Sprintf(msg, s.GracefulShutdownTime))

	select {
	case <-time.After(s.GracefulShutdownTime):
		return
	case s := <-ch:
		log.Fatalf("Received signal %s: terminating immediately", s)
	}
}

func (s *WysteriaServer) close() {
	s.database.Close()
	s.searchbase.Close()
	s.fromClients.Unsubscribe()
	s.fromServers.Unsubscribe()
	s.middleware.Close()
}

func (s *WysteriaServer) Connect() error {
	err := s.connect()
	if err != nil {
		return err
	}
	return nil
}

func (s *WysteriaServer) connect() error {
	msg := "Attempting connection to %s %s %s:%d"

	log.Println(fmt.Sprintf(msg, "database", Config.DatabaseSettings.Driver, Config.DatabaseSettings.Host, Config.DatabaseSettings.Port))
	database, err := wdb.Connect(Config.DatabaseSettings)
	if err != nil {
		return err
	}
	s.database = database

	log.Println(fmt.Sprintf(msg, "searchbase", Config.SearchbaseSettings.Driver, Config.SearchbaseSettings.Host, Config.SearchbaseSettings.Port))
	searchbase, err := wsb.Connect(Config.SearchbaseSettings)
	if err != nil {
		return err
	}
	s.searchbase = searchbase

	log.Println(fmt.Sprintf(msg, "middleware", Config.MiddlewareSettings.Driver, Config.MiddlewareSettings.Host, Config.MiddlewareSettings.Port))
	ware, err := wcm.Connect(Config.MiddlewareSettings)
	if err != nil {
		return err
	}
	s.middleware = ware

	log.Println(fmt.Sprintf("Subscribing to %s%s", s.SettingsMiddleware.RouteServer+">", server_queue))
	cSub, err := s.middleware.Subscribe(s.SettingsMiddleware.RouteServer+">", server_queue)
	if err != nil {
		return err
	}
	s.fromClients = cSub

	log.Println(fmt.Sprintf("Subscribing to %s%s", s.SettingsMiddleware.RouteInternalServer+">", server_queue))
	sSub, err := s.middleware.Subscribe(s.SettingsMiddleware.RouteInternalServer+">", server_queue)
	if err != nil {
		return err
	}
	s.fromServers = sSub

	return nil
}

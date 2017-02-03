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

	SettingsMiddleware *wcm.MiddlewareSettings
	SettingsDatabase   *wdb.DatabaseSettings
	SettingsSearchbase *wsb.SearchbaseSettings

	database   wdb.Database
	searchbase wsb.Searchbase
	middleware wcm.WysteriaMiddleware

	fromClients wcm.WysteriaSubscription
	fromServers wcm.WysteriaSubscription
}

// Main server loop
//  Endlessly listen for inbound messages and pass work to routines
//
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

func (s *WysteriaServer) pickClientHandler(subject string) func([]byte) ([]byte, error) {
	switch subject {

	// Search
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_COLLECTION:
		return s.handleFindCollection
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_ITEM:
		return s.handleFindItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_VERSION:
		return s.handleFindVersion
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_RESOURCE:
		return s.handleFindResource
	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_LINK:
		return s.handleFindLink

	case s.SettingsMiddleware.RouteServer + wyc.MSG_FIND_HIGHEST_VERSION:
		return s.handleFindHighestVersion

	// Create
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_COLLECTION:
		return s.handleCreateCollection
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_ITEM:
		return s.handleCreateItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_VERSION:
		return s.handleCreateVersion
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_RESOURCE:
		return s.handleCreateResource
	case s.SettingsMiddleware.RouteServer + wyc.MSG_CREATE_LINK:
		return s.handleCreateLink

	// Update
	case s.SettingsMiddleware.RouteServer + wyc.MSG_UPDATE_ITEM:
		return s.handleUpdateItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_UPDATE_VERSION:
		return s.handleUpdateVersion

	// Delete
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_COLLECTION:
		return s.handleDelCollection
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_ITEM:
		return s.handleDelItem
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_VERSION:
		return s.handleDelVersion
	case s.SettingsMiddleware.RouteServer + wyc.MSG_DELETE_RESOURCE:
		return s.handleDelResource

	// ?
	default:
		break
	}
	return nil
}

func (s *WysteriaServer) handleClientMessage(message *wcm.Message) {
	handler := s.pickClientHandler(message.Subject)
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

	log.Println(fmt.Sprintf(msg, "database", s.SettingsDatabase.Driver, s.SettingsDatabase.Host, s.SettingsDatabase.Port))
	database, err := wdb.Connect(s.SettingsDatabase)
	if err != nil {
		return err
	}
	s.database = database

	log.Println(fmt.Sprintf(msg, "searchbase", s.SettingsSearchbase.Driver, s.SettingsSearchbase.Host, s.SettingsSearchbase.Port))
	searchbase, err := wsb.Connect(s.SettingsSearchbase)
	if err != nil {
		return err
	}
	s.searchbase = searchbase

	log.Println(fmt.Sprintf(msg, "middleware", s.SettingsMiddleware.Driver, s.SettingsMiddleware.Host, s.SettingsMiddleware.Port))
	ware, err := wcm.Connect(s.SettingsMiddleware)
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

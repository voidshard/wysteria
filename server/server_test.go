package main

import (
	"testing"
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	"fmt"
)

func TestPickClientHandler(t *testing.T) {
	// arrange
	route := "foobar"
	s := WysteriaServer{
		SettingsMiddleware: &wcm.MiddlewareSettings{
			RouteServer: route,
		},
	}

	cases := []struct{
		Subject string
		Handler func([]byte) ([]byte, error)
	} {
		{wyc.MSG_FIND_COLLECTION, s.handleFindCollection,},
		{wyc.MSG_FIND_ITEM, s.handleFindItem,},
		{wyc.MSG_FIND_VERSION, s.handleFindVersion,},
		{wyc.MSG_FIND_RESOURCE, s.handleFindResource,},
		{wyc.MSG_FIND_LINK, s.handleFindLink,},
		{wyc.MSG_FIND_HIGHEST_VERSION, s.handleFindHighestVersion,},
		{wyc.MSG_CREATE_COLLECTION, s.handleCreateCollection,},
		{wyc.MSG_CREATE_ITEM, s.handleCreateItem,},
		{wyc.MSG_CREATE_VERSION, s.handleCreateVersion,},
		{wyc.MSG_CREATE_RESOURCE, s.handleCreateResource,},
		{wyc.MSG_CREATE_LINK, s.handleCreateLink,},
		{wyc.MSG_UPDATE_ITEM, s.handleUpdateItem,},
		{wyc.MSG_UPDATE_VERSION, s.handleUpdateVersion,},
		{wyc.MSG_DELETE_COLLECTION, s.handleDelCollection,},
		{wyc.MSG_DELETE_ITEM, s.handleDelItem,},
		{wyc.MSG_DELETE_VERSION, s.handleDelVersion,},
		{wyc.MSG_DELETE_RESOURCE, s.handleDelResource,},
		{"foobar garbage", nil,},
	}
	
	for _, tst := range cases {
		// act
		handler := s.pickClientHandler(route + tst.Subject)

		handler_address := fmt.Sprintf("%s", handler)
		expected_address := fmt.Sprintf("%s", tst.Handler)

		// assert
		if handler_address != expected_address {
			t.Error("Expected given subject", route + tst.Subject, tst.Handler, "got", tst.Subject)
		}
	}
}
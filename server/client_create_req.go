package main

import (
	"encoding/json"
	"errors"
	"log"
	wyc "github.com/voidshard/wysteria/common"
	"fmt"
)

func (s *WysteriaServer) handleCreateCollection(data []byte) ([]byte, error) {
	col := wyc.Collection{}
	err := json.Unmarshal(data, &col)
	if err != nil {
		return nil, err
	}

	if col.Name == "" { // Check required field
		return nil, errors.New("Name required for Collection")
	}

	col.Id = NewId()
	err = s.database.InsertCollection(col.Id, col)
	if err != nil {
		return nil, err
	}

	return json.Marshal(col)
}

func (s *WysteriaServer) handleCreateItem(data []byte) ([]byte, error) {
	i := wyc.Item{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	if i.Parent == "" || i.Variant == "" || i.ItemType == "" { // Check required fields
		return nil, errors.New("Parent, ItemType, Variant required for Item")
	}

	results, err := s.searchbase.QueryItem("", true, 0, wyc.QueryDesc{
		Parent: i.Parent,
		ItemType: i.ItemType,
		Variant: i.Variant,
	})
	if err != nil {
		return nil, err
	}
	if len(results) > 0 {
		return nil, errors.New(fmt.Sprintf("A child item of %s already exists with %s %s", i.Parent, i.ItemType, i.Variant))
	}

	i.Id = NewId()
	err = s.database.InsertItem(i.Id, i)
	if err != nil {
		return nil, err
	}

	err = s.searchbase.InsertItem(i.Id, i)
	if err != nil {
		log.Println(err)
	}

	return json.Marshal(i)
}

func (s *WysteriaServer) handleCreateVersion(data []byte) ([]byte, error) {
	i := wyc.Version{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	if i.Parent == "" {
		return nil, errors.New("Parent required for Version")
	}

	i.Id = NewId()
	number, err := s.database.InsertNextVersion(i.Id, i)
	if err != nil {
		return nil, err
	}

	i.Number = number
	err = s.searchbase.InsertVersion(i.Id, i)
	if err != nil {
		return nil, err
	}

	return json.Marshal(i)
}

func (s *WysteriaServer) handleCreateResource(data []byte) ([]byte, error) {
	i := wyc.Resource{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	if i.Name == "" || i.ResourceType == "" || i.Location == "" {
		return nil, errors.New("Name, ResourceType and Location required for Resource")
	}

	i.Id = NewId()
	err = s.database.InsertResource(i.Id, i)
	if err != nil {
		return nil, err
	}

	err = s.searchbase.InsertResource(i.Id, i)
	if err != nil {
		log.Println(err)
	}

	return json.Marshal(i)
}

func (s *WysteriaServer) handleCreateLink(data []byte) ([]byte, error) {
	i := wyc.Link{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	if i.Name == "" || i.Src == "" || i.Dst == "" {
		return nil, errors.New("Name, Src and Dst required for Link")
	}

	i.Id = NewId()
	err = s.database.InsertLink(i.Id, i)
	if err != nil {
		return nil, err
	}

	err = s.searchbase.InsertLink(i.Id, i)
	if err != nil {
		log.Println(err)
	}

	return json.Marshal(i)
}

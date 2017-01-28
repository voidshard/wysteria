package main

import (
	"encoding/json"
	wyc "github.com/voidshard/wysteria/common"
)

func (s *WysteriaServer) handleUpdateItem(data []byte) ([]byte, error) {
	i := wyc.Item{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	err = s.database.UpdateItem(i.Id, i)
	if err != nil {
		return nil, err
	}

	err = s.searchbase.UpdateItem(i.Id, i)
	if err != nil {
		return nil, err
	}

	return json.Marshal(i)
}

func (s *WysteriaServer) handleUpdateVersion(data []byte) ([]byte, error) {
	i := wyc.Version{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}

	err = s.database.UpdateVersion(i.Id, i)
	if err != nil {
		return nil, err
	}

	err = s.searchbase.UpdateVersion(i.Id, i)
	if err != nil {
		return nil, err
	}

	return json.Marshal(i)
}

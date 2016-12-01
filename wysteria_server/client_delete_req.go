package main

import (
	"encoding/json"
	"errors"
	wyc "wysteria/wysteria_common"
)

const (
	err_cannot_orphan = "Unable to delete obj while undeleted children exist"
	err_id_required   = "Unable to delete: Obj id not supplied"
)

func (s *WysteriaServer) handleDelCollection(data []byte) ([]byte, error) {
	i := wyc.Collection{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}
	if i.Id == "" {
		return nil, errors.New(err_id_required)
	}

	res, err := s.searchbase.QueryItem("", true, 0, wyc.QueryDesc{Parent: i.Id})
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return nil, errors.New(err_cannot_orphan)
	}

	err = s.database.DeleteCollection(i.Id)
	if err != nil {
		return nil, err
	}

	return wyc.WYSTERIA_SERVER_ACK, nil
}

func (s *WysteriaServer) handleDelItem(data []byte) ([]byte, error) {
	i := wyc.Item{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}
	if i.Id == "" {
		return nil, errors.New(err_id_required)
	}

	res, err := s.searchbase.QueryVersion("", true, 0, wyc.QueryDesc{Parent: i.Id})
	if len(res) > 0 {
		return nil, errors.New(err_cannot_orphan)
	}

	if i.Links != nil && len(i.Links) > 0 {
		err = s.database.DeleteLink(i.Links...)
		if err != nil {
			return nil, err
		}
		s.searchbase.DeleteLink(i.Links...)
	}

	err = s.database.DeleteItem(i.Id)
	if err != nil {
		return nil, err
	}
	s.searchbase.DeleteItem(i.Id)

	return wyc.WYSTERIA_SERVER_ACK, nil
}

func (s *WysteriaServer) handleDelVersion(data []byte) ([]byte, error) {
	i := wyc.Version{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}
	if i.Id == "" {
		return nil, errors.New(err_id_required)
	}

	res, err := s.searchbase.QueryFileResource("", true, 0, wyc.QueryDesc{Parent: i.Id})
	if err != nil {
		return nil, err
	}
	if len(res) > 0 {
		return nil, errors.New(err_cannot_orphan)
	}

	if i.Links != nil && len(i.Links) > 0 {
		err = s.database.DeleteLink(i.Links...)
		if err != nil {
			return nil, err
		}
		s.searchbase.DeleteLink(i.Links...)
	}

	err = s.database.DeleteVersion(i.Id)
	if err != nil {
		return nil, err
	}
	err = s.searchbase.DeleteVersion(i.Id)
	if err != nil {
		return nil, err
	}

	return wyc.WYSTERIA_SERVER_ACK, nil
}

func (s *WysteriaServer) handleDelFileResource(data []byte) ([]byte, error) {
	i := wyc.FileResource{}
	err := json.Unmarshal(data, &i)
	if err != nil {
		return nil, err
	}
	if i.Id == "" {
		return nil, errors.New(err_id_required)
	}

	err = s.database.DeleteCollection(i.Id)
	if err != nil {
		return nil, err
	}
	s.searchbase.DeleteFileResource(i.Id)

	return wyc.WYSTERIA_SERVER_ACK, nil
}

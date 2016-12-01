package main

import (
	wyc "wysteria/wysteria_common"
	"encoding/json"
	"errors"
	"fmt"
)

func (s *WysteriaServer) findHighestVersion(parentId string) (wyc.Version, error) {
	// Grab the highest ten in case we are currently deleting 9 Versions (for some reason ..)
	ids, err := s.searchbase.QueryVersion("Number", false, 10, wyc.QueryDesc{Parent: parentId})
	if err != nil {
		return wyc.Version{}, err
	}

	if len(ids) == 0 {
		return wyc.Version{}, errors.New("No versions found")
	}

	// Edge case; A version could be removed in the DB but still be in the SB. We need to handle that.
	// We expect that in almost all cases we'll grab the first (highest) version from the db, but in the
	// event that the highest has been removed from the DB but not the searchbase yet, we'll go down our
	// list highest to lowest and return the highest non deleted (in the db) version
	for _, id := range(ids) {
		last, err := s.database.RetrieveVersion(id)
		if err != nil {
			// in this case, the Version has been deleted in the DB but not yet in the searchbase.
			// Queue (another?) SearchBase delete and check the next highest version number
			delete_id := id
			go func() {
				s.searchbase.DeleteVersion(delete_id)
			}()
		} else {
			if len(last) > 0 {
				return last[0], nil
			}
		}
	}

	// We can arrive here if the searchbase and the database are currently too far out of sync.
	// Because deletes for the searchbase have been queued at this point, this should come right - so
	// long as our searchbase isn't currently dead in the water.
	return wyc.Version{}, errors.New("Unable to establish highest version")
}

func (s *WysteriaServer) handleFindHighestVersion(data []byte) ([]byte, error) {
	ver := wyc.Version{}
	err := json.Unmarshal(data, &ver)
	if err != nil {
		return nil, err
	}

	fver, err := s.findHighestVersion(ver.Parent)
	if err != nil {
		return nil, err
	}
	if fver.Parent == "" {
		return nil, errors.New(fmt.Sprintf("Unable to find version with parent %s", ver.Parent))
	}

	return json.Marshal(fver)
}

func (s *WysteriaServer) handleFindCollection(data []byte) ([]byte, error) {
	qs := []wyc.QueryDesc{}
	err := json.Unmarshal(data, &qs)
	if err != nil {
		return nil, err
	}

	ids := []string{}
	names := []string{}
	for _, qd := range qs {
		if qd.Id != "" {
			ids = append(ids, qd.Id)
		}
		if qd.Name != "" {
			names = append(names, qd.Name)
		}
	}

	results := []wyc.Collection{}
	if len(ids) > 0 {
		items, err := s.database.RetrieveCollection(ids...)
		if err == nil { // Do not exit if/when we've been given invalid id(s)
			results = append(results, items...)
		}
	}
	if len(names) > 0 {
		items, err := s.database.RetrieveCollectionByName(names...)
		if err != nil {
			return nil, err
		}
		results = append(results, items...)
	}
	return json.Marshal(results)
}

func (s *WysteriaServer) handleFindItem(data []byte) ([]byte, error) {
	qs := []wyc.QueryDesc{}
	err := json.Unmarshal(data, &qs)
	if err != nil {
		return nil, err
	}

	ids, err := s.searchbase.QueryItem("", true, 0, qs...)
	if err != nil {
		return nil, err
	}

	items, err := s.database.RetrieveItem(ids...)
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

func (s *WysteriaServer) handleFindVersion(data []byte) ([]byte, error) {
	qs := []wyc.QueryDesc{}
	err := json.Unmarshal(data, &qs)
	if err != nil {
		return nil, err
	}

	ids, err := s.searchbase.QueryVersion("", true, 0, qs...)
	if err != nil {
		return nil, err
	}

	items, err := s.database.RetrieveVersion(ids...)
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

func (s *WysteriaServer) handleFindFileResource(data []byte) ([]byte, error) {
	qs := []wyc.QueryDesc{}
	err := json.Unmarshal(data, &qs)
	if err != nil {
		return nil, err
	}

	ids, err := s.searchbase.QueryFileResource("", true, 0, qs...)
	if err != nil {
		return nil, err
	}

	items, err := s.database.RetrieveFileResource(ids...)
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

func (s *WysteriaServer) handleFindLink(data []byte) ([]byte, error) {
	qs := []wyc.QueryDesc{}
	err := json.Unmarshal(data, &qs)
	if err != nil {
		return nil, err
	}

	ids, err := s.searchbase.QueryLink("", true, 0, qs...)
	if err != nil {
		return nil, err
	}

	items, err := s.database.RetrieveLink(ids...)
	if err != nil {
		return nil, err
	}

	return json.Marshal(items)
}

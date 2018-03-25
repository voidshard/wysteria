package main

import (
	wyc "github.com/voidshard/wysteria/common"
)


// Return if the given key is in the given list
func ListContains(key string, values []string) bool {
	for _, v := range values {
		if v == key {
			return true
		}
	}
	return false
}

// Divide the search queries into a list of results to get by ID and a list of queries we need to run
//
func ExtractIdQueries(qs []*wyc.QueryDesc) ([]string, []*wyc.QueryDesc) {
	ids := []string{}
	queries := []*wyc.QueryDesc{}

	for _, q := range qs {
		if IsIdOnlyQuery(q) {
			ids = append(ids, q.Id)
		} else {
			queries = append(queries, q)
		}
	}

	return ids, queries
}

// Checks if ID is the *only* field the user is searching for. If so, we can simply get it from the DB.
//
func IsIdOnlyQuery(q *wyc.QueryDesc) bool {
	if q.Id == "" { // no id is set, so it can't be an id only query
		return false
	}
	if q.Facets != nil {
		return false
	}
	if q.VersionNumber > 0 {
		return false
	}
	for _, s := range []string{q.Parent, q.ItemType, q.Variant, q.Name, q.ResourceType, q.Location, q.LinkSrc, q.LinkDst} {
		if s != "" {
			return false
		}
	}
	return true
}

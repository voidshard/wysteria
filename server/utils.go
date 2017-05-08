package main

import "gopkg.in/mgo.v2/bson"

// Create a new ID string at random
func NewId() string {
	return bson.NewObjectId().Hex()
}

// Return if the given key is in the given list
func ListContains(key string, values []string) bool {
	for _, v := range values {
		if v == key {
			return true
		}
	}
	return false
}

package main

import "gopkg.in/mgo.v2/bson"

// Create a new ID string at random
func NewId() string {
	return bson.NewObjectId().Hex()
}

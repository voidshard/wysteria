package main

import "gopkg.in/mgo.v2/bson"

func NewId() string {
	return bson.NewObjectId().Hex()
}

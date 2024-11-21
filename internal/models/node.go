package models

import (
	"time"
)

func NewNode() *Node {
	return &Node{
		ID: int(time.Now().UnixNano() % 10000), // Simple ID generation for demo
		Location: Location{
			Lat: 0,
			Lon: 0,
		},
		MediaFiles: make([]MediaFile, 0),
		EntryCondition: &Condition{
			Type:    "q&a",
			Strict:  false,
			Options: make([]string, 0),
			Hints:   make([]string, 0),
		},
		ExitCondition: &Condition{
			Type:    "q&a",
			Strict:  false,
			Options: make([]string, 0),
			Hints:   make([]string, 0),
		},
	}
}

func (t *Tour) GetNode(id int) *Node {
	for i := range t.Nodes {
		if t.Nodes[i].ID == id {
			return &t.Nodes[i]
		}
	}
	return nil
}

package message

import (
	"sync"
)

const (
	ServerMessageActionStop = iota
	ServerMessageActionPause = iota
	ServerMessageActionResume = iota
)

type ServerMessage struct {
	Uuid string
	Action int
}

type TraderStore struct {
	traders map[string] chan ServerMessage
	mu sync.RWMutex
}


func (t *TraderStore) Add(uuid string)  chan ServerMessage {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.traders[uuid]; ok {
		panic("Trader channel already exists")
	}
	t.traders[uuid] = make(chan ServerMessage)
	return t.traders[uuid]
}


func (t *TraderStore) Del(uuid string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if traderCh, ok := t.traders[uuid]; ok {
		traderCh <- ServerMessage{uuid, ServerMessageActionStop}
		close(t.traders[uuid])
		delete(t.traders, uuid)
	}
}


func NewTraderStore() *TraderStore {
	return &TraderStore{traders: make(map[string] chan ServerMessage)}
}
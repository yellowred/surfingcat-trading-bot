package message

import (
	"errors"
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

var traders map[string] chan ServerMessage

func SeverMessageToTrader(uuid string, msg int) (bool, error) {
	if traderCh, ok := traders[uuid]; ok {
		traderCh <- ServerMessage{uuid, msg}
		return ok, nil
	} else {
		return false, errors.New("Trader chan not found.")
	}
}

func NewChannelToTrader(uuid string) chan ServerMessage {
	if traders == nil {
		traders = make(map[string] chan ServerMessage)
	}
	if _, ok := traders[uuid]; ok {
		panic("Trader channel already exists")
	}
	traders[uuid] = make(chan ServerMessage)
	return traders[uuid]
}

func StopTrader(uuid string) error {
	if ok, err := SeverMessageToTrader(uuid, ServerMessageActionStop); ok {
		close(traders[uuid])
		return nil
	} else {
		return err
	}
}
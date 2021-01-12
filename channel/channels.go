package channel

import (
	"context"
	"fmt"
	"math/rand"
)

type (
	Channels struct {
		conf     Config
		channels map[string]*Channel
	}
	Config struct {
		VTEndpoint     string
		OutputEndpoint string
	}
)

func New() *Channels {
	ch := &Channels{
		conf: Config{
			VTEndpoint:     "http://localhost:7071",
			OutputEndpoint: "http://localhost:7072",
		},
		channels: make(map[string]*Channel),
	}
	return ch
}

func (ch *Channels) Get(ctx context.Context, channelID string) (*Channel, error) {
	return ch.channels[channelID], nil
}

func (ch *Channels) New(ctx context.Context, newCh NewChannelStruct) (*Channel, error) {
	channel := &Channel{
		Name:        newCh.Name,
		Description: newCh.Description,
		ChannelType: newCh.ChannelType,
		OriginURL:   newCh.OriginURL,
		SlateURL:    newCh.SlateURL,
		Renditions:  newCh.Renditions,
		Archive:     newCh.Archive,
		DVR:         newCh.DVR,
		OriginOnly:  newCh.OriginOnly,
	}
	channel.Status = "pending"

	for {
		channel.ID = randString()
		_, exists := ch.channels[channel.ID]
		if !exists {
			break
		}
	}

	ch.channels[channel.ID] = channel

	return channel, nil
}

func (ch *Channels) Delete(channelID string) error {
	if _, ok := ch.channels[channelID]; ok {
		err := ch.channels[channelID].Stop()
		if err != nil {
			return fmt.Errorf("failed to delete channel: %w", err)
		}
		delete(ch.channels, channelID)
	}
	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randString() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

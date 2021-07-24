// Simple example embedding and extending Gate.
package main

import (
	"encoding/json"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	. "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/component/codec/legacy"
	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/runtime/event"
	"go.minekube.com/gate/pkg/util/favicon"
	gateUUID "go.minekube.com/gate/pkg/util/uuid"
	"os"
	"time"
)

func main() {
	// Add our "plug-in" to be initialized on Gate start.
	proxy.Plugins = append(proxy.Plugins, proxy.Plugin{
		Name: "SimpleProxy",
		Init: func(proxy *proxy.Proxy) error {
			return newSimpleProxy(proxy).init()
		},
	})

	// Execute Gate entrypoint and block until shutdown.
	// We could also run gate.Start if we don't need Gate's command-line.
	gate.Execute()
}

// SimpleProxy is a simple proxy that adds a `/broadcast` command
// and sends a message on server switch.
type SimpleProxy struct {
	*proxy.Proxy
	legacyCodec *legacy.Legacy
}

func newSimpleProxy(proxy *proxy.Proxy) *SimpleProxy {

	return &SimpleProxy{
		Proxy:       proxy,
		legacyCodec: &legacy.Legacy{Char: legacy.AmpersandChar},
	}
}

// initialize our sample proxy
func (p *SimpleProxy) init() error {
	return p.registerSubscribers()
}

type ServerStatus struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	Favicon string
	//favicon ignored
}

// Register event subscribers
func (p *SimpleProxy) registerSubscribers() error {
	// Change the MOTD response.
	//motd := &Text{Content: "Simple Proxy!\nJoin and test me."}
	p.Event().Subscribe(&proxy.PingEvent{}, 0, func(ev event.Event) {
		e := ev.(*proxy.PingEvent)

		serverPing := e.Ping()
		resp, _, err := bot.PingAndListTimeout(p.Servers()[0].ServerInfo().Addr().String(), time.Second*1)

		if err != nil {
			serverPing.Description = &Text{Content: "Failed to ping server"}
			return
		}
		status := ServerStatus{}
		err = json.Unmarshal(resp, &status)
		if err != nil {
			fmt.Print("unmarshal resp fail:", err)
			os.Exit(1)
		}
		serverPing.Description = &Text{Content: status.Description.Text}
		//p.Description = &Text{Content: "aaaaa"}
		serverPing.Players.Max = status.Players.Max
		serverPing.Players.Online = status.Players.Online
		for _, player := range status.Players.Sample {
			//print(player.Name)
			serverPing.Players.Sample = append(serverPing.Players.Sample, ping.SamplePlayer{
				ID:   gateUUID.UUID(player.ID),
				Name: player.Name,
			})
		}

		serverPing.Favicon = favicon.Favicon(status.Favicon)
	})

	return nil
}

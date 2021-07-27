// Simple example embedding and extending Gate.
package main

import (
	"encoding/json"
	"fmt"
	qqsdk "github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	cq "github.com/Mrs4s/go-cqhttp/app"
	"github.com/Mrs4s/go-cqhttp/coolq"
	cqConfig "github.com/Mrs4s/go-cqhttp/global/config"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	"github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	. "go.minekube.com/common/minecraft/component"
	"go.minekube.com/common/minecraft/component/codec/legacy"
	"go.minekube.com/gate/cmd/gate"
	"go.minekube.com/gate/pkg/edition/java/ping"
	"go.minekube.com/gate/pkg/edition/java/proxy"
	"go.minekube.com/gate/pkg/runtime/event"
	"go.minekube.com/gate/pkg/util/favicon"
	gateUUID "go.minekube.com/gate/pkg/util/uuid"
	"os"
	"strings"
	"time"
)

type MainConfig struct {
	Controller Config
}
type Config struct {
	QQ struct {
		Enable       bool
		Group        int64
		Chat         bool
		Notification struct {
			Online bool
		}
	}
}

var qqclient *qqsdk.QQClient
var config Config

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
	//gate.Execute()

	log.SetFormatter(&log.TextFormatter{ForceColors: true})
	log.SetOutput(colorable.NewColorableStdout())

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.WithError(err).Fatal("Fatal error read config file")
	}

	configWrapper := MainConfig{}
	err = viper.Unmarshal(&configWrapper)
	if err != nil {
		log.WithError(err).Fatal("unable to decode config into struct")
	}

	config = configWrapper.Controller

	if config.QQ.Enable {
		conf := cq.InitConfig(false, cqConfig.DefaultConfigFile)
		qqclient = cq.InitClient(nil)
		coolq.NewQQBot(qqclient, conf)
	}

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
		resp, _, err := bot.PingAndListTimeout(p.Servers()[0].ServerInfo().Addr().String(), time.Millisecond*500)

		if err != nil {
			log.WithField("err", err).Info("ping fail")
			serverPing.Description = &Text{Content: "服務器沒有開機，請先連線來觸發啟動"}
			serverPing.Players.Max = 0
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
	if config.QQ.Enable {

		if config.QQ.Chat {
			qqclient.OnGroupMessage(func(client *qqsdk.QQClient, groupMessage *message.GroupMessage) {
				if groupMessage.GroupCode == config.QQ.Group {
					senderName := groupMessage.Sender.CardName
					if strings.TrimSpace(senderName) == "" {
						senderName = groupMessage.Sender.Nickname
					}
					for _, player := range p.Players() {
						_ = player.SendMessage(&Text{

							Content: fmt.Sprintf("[QQ][%s]%s", senderName, groupMessage.ToString()),
						})
					}
				}

			})

			p.Event().Subscribe(&proxy.PlayerChatEvent{}, 0, func(ev event.Event) {
				e := ev.(*proxy.PlayerChatEvent)
				qqclient.SendGroupMessage(config.QQ.Group, &message.SendingMessage{Elements: []message.IMessageElement{
					&message.TextElement{
						Content: fmt.Sprintf("[MC][%s]%s", e.Player().Username(), e.Message()),
					},
				}})
			})
		}

		if config.QQ.Notification.Online {
			p.Event().Subscribe(&proxy.ServerPostConnectEvent{}, 0, func(ev event.Event) {
				e := ev.(*proxy.ServerPostConnectEvent)
				qqclient.SendGroupMessage(config.QQ.Group, &message.SendingMessage{Elements: []message.IMessageElement{
					&message.TextElement{
						Content: fmt.Sprintf("[MC]%s 進入遊戲", e.Player().Username()),
					},
				}})

			})

			p.Event().Subscribe(&proxy.DisconnectEvent{}, 0, func(ev event.Event) {
				e := ev.(*proxy.DisconnectEvent)
				qqclient.SendGroupMessage(config.QQ.Group, &message.SendingMessage{Elements: []message.IMessageElement{
					&message.TextElement{
						Content: fmt.Sprintf("[MC]%s 離開遊戲", e.Player().Username()),
					},
				}})

			})
		}
	}

	//p.Event().Subscribe(&proxy.KickedFromServerEvent{}, 0, func(ev event.Event) {
	//	e := ev.(*proxy.KickedFromServerEvent)
	//	//print(e)
	//	//e.Result()
	//
	//
	//	 if result,ok := e.Result().(*proxy.DisconnectPlayerKickResult); ok {
	//	 	if reason,ok := result.Reason.(*Text);ok {
	//	 		log.WithField("reason",reason).Info("Player unable to connect to to server")
	//	 		if strings.Contains(reason.Content,"Unable to connect to"){
	//				reason.Content = "服務器沒有開機，請稍等5分鐘"
	//			}
	//	 		//print("hello how are you\n")
	//
	//		}
	//
	//	 }
	//})

	return nil
}

module github.com/LaysDragon/MinecraftOopsOnlineGate

go 1.14

require (
	github.com/Mrs4s/MiraiGo v0.0.0-20210525010101-8f0cd9494d64
	github.com/Mrs4s/go-cqhttp v0.9.40
	github.com/Tnze/go-mc v1.17.0
	github.com/google/uuid v1.3.0
	github.com/mattn/go-colorable v0.1.8
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/viper v1.8.1
	go.minekube.com/brigodier v0.0.0-20210619074847-0327262d340c // indirect
	go.minekube.com/common v0.0.0-20210614170949-47d30d7729ca
	go.minekube.com/gate v0.13.0
)

replace go.minekube.com/gate => github.com/LaysDragon/gate v0.13.1-0.20210727074757-973456d287bb

replace github.com/Mrs4s/go-cqhttp => github.com/LaysDragon/go-cqhttp v1.0.0-beta4.0.20210725140814-722fdd4f3eb7

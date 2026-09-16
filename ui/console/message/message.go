package message

import "netfs/api"

type ResizeMsg struct {
	Width  int
	Height int
}

type TriggerMsg struct{}

type RefreshedHost struct {
	Host  api.Host
	Alive bool
}

type RefreshStateMsg struct {
	Hosts []RefreshedHost
}

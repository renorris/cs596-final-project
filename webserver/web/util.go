package web

type AlertMsg struct {
	AlertMsg string
}

func NewAlertMsg(msg string) AlertMsg {
	return AlertMsg{AlertMsg: msg}
}

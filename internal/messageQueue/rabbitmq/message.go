package rabbitmq

import (
	"SnapLink/internal/model"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"net/http"
)

type AccessLogMessage struct {
	Info     model.Redirect `json:"info"`
	Datetime string         `json:"datetime"`
	Header   http.Header    `json:"header"`
	IP       string         `json:"ip"`
	UID      string         `json:"uid"`
}

// NewAccessLogMessage 生成访问日志
func NewAccessLogMessage(info model.Redirect, header http.Header, ip, uid, datetime string) *message.Message {
	jsonByes, _ := json.Marshal(AccessLogMessage{
		Info:     info,
		Header:   header,
		Datetime: datetime,
		IP:       ip,
		UID:      uid,
	})
	return message.NewMessage(watermill.NewUUID(), jsonByes)
}

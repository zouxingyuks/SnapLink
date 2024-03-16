package messageQueue

import (
	"SnapLink/internal/model"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"net/http"
)

type AccessLogMessage struct {
	Info   *model.Redirect `json:"info"`
	Header *http.Header    `json:"header"`
	Uri    string          `json:"uri"`
}

// NewAccessLogMessage 生成访问日志
func NewAccessLogMessage(info *model.Redirect, header *http.Header, uri string) *message.Message {
	jsonByes, _ := json.Marshal(AccessLogMessage{
		Info:   info,
		Header: header,
		Uri:    uri,
	})
	return message.NewMessage(watermill.NewUUID(), jsonByes)
}

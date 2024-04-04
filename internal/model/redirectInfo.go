package model

import (
	"fmt"
)

type Redirect struct {
	ID          int    `gorm:"column:id;primary_key;auto_increment;comment:'主键';not null" json:"-"`
	Uri         string `gorm:"type:varchar(10);column:uri;commit:'生成短链接的uri';not null;index:idx_uri" json:"uri"`
	Gid         string `gorm:"column:gid;comment:'组id';not null" json:"gid,omitempty"`
	OriginalURL string `gorm:"type:nvarchar(255);column:original_URL;comment:'原始链接';not null;" json:"originalURL,omitempty"`
	VaildDate   string `gorm:"-" json:"vaildDate,omitempty"`
	VaildType   int    `gorm:"-" json:"vaildType,omitempty"`
}

func (r Redirect) TName() string {
	id := hash(r.Uri)
	return fmt.Sprintf("%s-%d", RedirectPrefix, id%RedirectShardingNum)

}

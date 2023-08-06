package model

const (
	CMD_SINGLE_MSG = 10
	CMD_ROOM_MSG   = 11
	CMD_HEART      = 0
)

type Message struct {
	Id      int64  `json:"id,omitempty" form:"id"`           //消息ID
	Userid  int64  `json:"userid,omitempty" form:"userid"`   //谁发的
	Cmd     int    `json:"cmd,omitempty" form:"cmd"`         //群聊还是私聊
	Dstid   int64  `json:"dstid,omitempty" form:"dstid"`     //对端用户ID/群ID
	Media   int    `json:"media,omitempty" form:"media"`     //消息按照什么样式展示
	Content string `json:"content,omitempty" form:"content"` //消息的内容
	Pic     string `json:"pic,omitempty" form:"pic"`         //预览图片
	Url     string `json:"url,omitempty" form:"url"`         //服务的URL
	Memo    string `json:"memo,omitempty" form:"memo"`       //简单描述
	Amount  int    `json:"amount,omitempty" form:"amount"`   //其他和数字相关的
}

type HistoryMessage struct {
	Id        int64  `json:"id,omitempty" form:"id" xorm:"int"`                    //消息ID
	Userid    int64  `json:"userid,omitempty" form:"userid" xorm:"int"`            //谁发的
	Cmd       int    `json:"cmd,omitempty" form:"cmd" xorm:"int"`                  //群聊还是私聊
	Dstid     int64  `json:"dstid,omitempty" form:"dstid" xorm:"int"`              //对端用户ID/群ID
	Media     int    `json:"media,omitempty" form:"media" xorm:"int"`              //消息按照什么样式展示
	Content   string `json:"content,omitempty" form:"content" xorm:"varchar(255)"` //消息的内容
	Pic       string `json:"pic,omitempty" form:"pic" xorm:"varchar(255)"`         //预览图片
	Url       string `json:"url,omitempty" form:"url" xorm:"varchar(255)"`         //服务的URL
	Memo      string `json:"memo,omitempty" form:"memo" xorm:"varchar(255)"`       //简单描述
	Amount    int    `json:"amount,omitempty" form:"amount" xorm:"bigint"`         //其他和数字相关的
	Timestamp int64  `json:"timestamp" xorm:"bigint"`
}

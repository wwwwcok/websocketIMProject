package model

import (
	"fmt"
	"time"
)

//好友和群都存在这个表里面
//可根据具体业务做拆分
type Contact struct {
	Id int64 `xorm:"pk autoincr bigint(20)" form:"id" json:"id"`
	//谁的10000
	Ownerid int64 `xorm:"bigint(20)" form:"ownerid" json:"ownerid"` // 记录是谁的
	//对端,10001
	Dstobj int64 `xorm:"bigint(20)" form:"dstobj" json:"dstobj"` // 对端信息
	//
	Cate int    `xorm:"int(11)" form:"cate" json:"cate"`      // 什么类型
	Memo string `xorm:"varchar(120)" form:"memo" json:"memo"` // 备注
	//
	Createat time.Time `xorm:"datetime" form:"createat" json:"createat"` // 创建时间
}

const (
	CONCAT_CATE_USER     = 0x01
	CONCAT_CATE_COMUNITY = 0x02
)

type ContactArg struct {
	PageArg
	Userid int64 `json:"userid" form:"userid"`
	Dstid  int64 `json:"dstid" form:"dstid"`
}

type PageArg struct {
	//从哪页开始
	Pagefrom int `json:"pagefrom" form:"pagefrom"`
	//每页大小
	Pagesize int `json:"pagesize" form:"pagesize"`
	//关键词
	Kword string `json:"kword" form:"kword"`
	//asc：“id”  id asc
	Asc  string `json:"asc" form:"asc"`
	Desc string `json:"desc" form:"desc"`
	//
	Name string `json:"name" form:"name"`
	//
	Userid int64 `json:"userid" form:"userid"`
	//dstid
	Dstid int64 `json:"dstid" form:"dstid"`
	//时间点1
	Datefrom time.Time `json:"datafrom" form:"datafrom"`
	//时间点2
	Dateto time.Time `json:"dateto" form:"dateto"`
	//
	Total int64 `json:"total" form:"total"`
}

func (p *PageArg) GetPageSize() int {
	if p.Pagesize == 0 {
		return 100
	} else {
		return p.Pagesize
	}

}
func (p *PageArg) GetPageFrom() int {
	if p.Pagefrom < 0 {
		return 0
	} else {
		return p.Pagefrom
	}
}

func (p *PageArg) GetOrderBy() string {
	if len(p.Asc) > 0 {
		return fmt.Sprintf(" %s asc", p.Asc)
	} else if len(p.Desc) > 0 {
		return fmt.Sprintf(" %s desc", p.Desc)
	} else {
		return ""
	}
}

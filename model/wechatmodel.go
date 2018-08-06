package model

import (
	"time"
)

// ORM映射中的MODEL对象
// **********************************************************************
type WechatData struct {
	Id         int64     `xorm:"id"` // XORM自动自增长
	TagId      string     `xorm:"tag_id"`
	TagName    string    `xorm:"tag_name"`
	Url        string    `xorm:"url"`
	Title      string    `xorm:"title"`
	Content    string    `xorm:"content"`
	Image      string    `xorm:"image"`
	Source     string    `xorm:"source"`
	CreateTime time.Time `xorm:"create_time"`
	UpdateTime time.Time `xorm:"update_time"`
	DeleteTime time.Time `xorm:"delete_time"`
}

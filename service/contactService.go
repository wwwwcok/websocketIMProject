package service

import (
	"errors"
	"time"
	"websocketIMProject/model"
	"websocketIMProject/tool"
)

type ContactService struct {
}

func (m *ContactService) AddFriend(userid, dstid int64) error {
	//判断是不是自己
	if userid == dstid {
		return errors.New("不能添加自己为好友")
	}
	//判断是否已经添加了好友
	tmp := &model.Contact{}
	dbEngine := tool.XormMysqlEngine
	b, err := dbEngine.Where("ownerid = ?", userid).And("dstobj = ?", dstid).And("cate = ?", model.CONCAT_CATE_USER).Get(tmp)
	if b {
		return errors.New("该用户已经添加好友")
	}
	if err != nil {
		return err
	}
	//事务相关
	session := dbEngine.NewSession()
	//事务开始
	err = session.Begin()
	if err != nil {
		return err
	}
	_, serr1 := dbEngine.InsertOne(&model.Contact{
		Ownerid:  userid,
		Dstobj:   dstid,
		Cate:     model.CONCAT_CATE_USER,
		Createat: time.Now(),
	})

	_, serr2 := dbEngine.InsertOne(&model.Contact{
		Ownerid:  dstid,
		Dstobj:   userid,
		Cate:     model.CONCAT_CATE_USER,
		Createat: time.Now(),
	})

	if serr2 != nil && serr1 != nil {
		session.Commit()
	} else {
		session.Rollback()
		if serr1 != nil {
			return serr1
		} else {
			return serr2
		}
	}

	return err
}

func (m *ContactService) SearchFriend(userId int64) ([]model.User, error) {
	dbEngine := tool.XormMysqlEngine
	tmp := make([]model.Contact, 0)
	err := dbEngine.Cols("dstobj").Where("ownerid = ? and cate = ?", userId, model.CONCAT_CATE_USER).Find(&tmp)
	if err != nil {
		return nil, err
	}
	dstIds := make([]int64, 0)

	for _, v := range tmp {
		dstIds = append(dstIds, v.Dstobj)
	}
	contacts := make([]model.User, 0)
	err = dbEngine.In("id", dstIds).Find(&contacts)
	return contacts, err
}

func (m *ContactService) SearchComunity(userId int64) ([]model.Community, error) {
	dbEngine := tool.XormMysqlEngine
	tmp := make([]model.Contact, 0)
	err := dbEngine.Cols("dstobj").Where("ownerid = ? and cate = ?", userId, model.CONCAT_CATE_COMUNITY).Find(&tmp)
	if err != nil {
		return nil, err
	}
	dstIds := make([]int64, 0)

	for _, v := range tmp {
		dstIds = append(dstIds, v.Dstobj)
	}
	contacts := make([]model.Community, 0)
	err = dbEngine.In("id", dstIds).Find(&contacts)
	return contacts, err
}
func (m *ContactService) SearchComunityIds(userId int64) []int64 {
	dbEngine := tool.XormMysqlEngine
	tmp := make([]model.Contact, 0)
	err := dbEngine.Where("ownerid = ? and cate = ?", userId, model.CONCAT_CATE_COMUNITY).Find(&tmp)
	if err != nil {
		return nil
	}
	comIds := make([]int64, 0)
	for _, v := range tmp {
		comIds = append(comIds, v.Dstobj)
	}

	return comIds

}

func (m *ContactService) JoinCommunity(userid, comId int64) error {
	//判断是否已经添加了好友
	tmp := &model.Contact{
		Ownerid: userid,
		Dstobj:  comId,
		Cate:    model.CONCAT_CATE_COMUNITY,
	}
	dbEngine := tool.XormMysqlEngine
	b, err := dbEngine.Get(tmp)
	if b {
		return errors.New("该用户已经加群")
	}
	if err != nil {
		return err
	}
	_, err = dbEngine.InsertOne(tmp)
	return err
}

func (m *ContactService) CreateCommunity(comm model.Community) (ret model.Community, err error) {
	if len(comm.Name) == 0 {
		return ret, errors.New("无群名称")
	}
	if comm.Ownerid == 0 {
		return ret, errors.New("请登录")
	}
	dbEngine := tool.XormMysqlEngine

	num, err := dbEngine.Where("ownerid = ? ", comm.Ownerid).Count(&comm)
	if err != nil {
		return ret, err
	}

	if num > 5 {
		return ret, errors.New("该用户创建群达到上限5")
	}
	comm.Createat = time.Now()
	//开启事务，为了同时插入一条数据到两张表
	//先生成一个单独的事务，如果不这样做，每一个插入或者where等行为都是一个不同的事务
	session := dbEngine.NewSession()
	//此单独的事务进行开启和关闭
	session.Begin()
	//插入到群组表
	_, err = session.InsertOne(&comm)
	if err != nil {
		session.Rollback()
		return ret, errors.New("插入到群组表时失败")
	}
	//插入到联系人表
	_, err = session.InsertOne(model.Contact{
		Ownerid:  comm.Ownerid,
		Dstobj:   comm.Id,
		Cate:     model.CONCAT_CATE_COMUNITY,
		Createat: comm.Createat,
	})
	if err != nil {
		session.Rollback()
		return ret, errors.New("插入到联系人表时失败")
	}
	err = session.Commit()

	ret = model.Community{
		Ownerid: comm.Ownerid,
	}
	return ret, err
}

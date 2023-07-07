package service

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
	"websocketIMProject/model"
	"websocketIMProject/tool"
)

type UserService struct {
}

func (m *UserService) Register(
	mobile,
	plainpwd,
	avatar,
	sex,
	nickname string) (user model.User, err error) {
	//根据手机号查看用户是否注册
	tmp := model.User{}
	dbEngine := tool.XormMysqlEngine
	exist, err := dbEngine.Where("mobile = ?", mobile).Get(&tmp)
	if err != nil {
		return tmp, err
	}
	//存在则返回提示信息
	if exist {
		return tmp, errors.New("该手机号已经存在")
	}
	//不存在则进行注册，拼接数据
	//密码哈希
	tmp.Salt = fmt.Sprintln(rand.Int31n(1000))
	password := tool.MD5Encode(plainpwd + tmp.Salt)
	fmt.Println("这是注册时候的手机号和密码--------------", mobile, plainpwd, tmp.Salt)
	tmp.Avatar = avatar
	tmp.Sex = sex
	tmp.Nickname = nickname
	tmp.Passwd = password
	tmp.Mobile = mobile
	tmp.Createat = time.Now()
	_, err = dbEngine.InsertOne(&tmp)

	return tmp, err
}

func (m *UserService) Login(
	mobile, plainpwd string) (user model.User, err error) {
	tmp := model.User{}
	dbEngine := tool.XormMysqlEngine
	//根据手机号查找密码
	exist, err := dbEngine.Where("mobile = ?", mobile).Get(&tmp)
	if err != nil {
		return tmp, err
	}
	//存在则返回提示信息
	if !exist {
		return tmp, errors.New("当前手机号不存在，请注册")
	}
	//密码验证
	if !tool.ValidatePasswd(plainpwd, tmp.Salt, tmp.Passwd) {
		return tmp, errors.New("密码错误")
	}

	//刷新token
	str := fmt.Sprintf("%d", time.Now().Unix())
	token := tool.MD5Encode(str)
	tmp.Token = token
	_, err = dbEngine.Cols("token").Where("id = ?", tmp.Id).Update(&tmp)
	if err != nil {
		return tmp, err
	}

	return tmp, err
}

//查找某个用户
func (m *UserService) Find(userId int64) model.User {
	tmp := model.User{}
	dbEngine := tool.XormMysqlEngine
	_, err := dbEngine.Where("id = ?", userId).Get(&tmp)
	if err != nil {
		fmt.Sprintln(err)
	}
	return tmp

}

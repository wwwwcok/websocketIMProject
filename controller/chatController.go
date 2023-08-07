package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"log"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
	"websocketIMProject/model"
	"websocketIMProject/service"
	"websocketIMProject/tool"
)

type Node struct {
	Conn *websocket.Conn
	//并行转串行,
	DataQueue     chan []byte
	GroupSets     set.Interface
	LastHeartbeat int64
	//在线状态，暂时未用
	Status bool
}

var (
	clientMap map[int64]*Node = make(map[int64]*Node, 0)
	rwlocker  sync.RWMutex
)

const maxIdleConnectionTime = time.Minute * 30 //半个小时

const maxIdleduration = time.Hour * 24 * 90 //90天

type ChatController struct {
}

func (m *ChatController) Router(Engine *gin.Engine) {
	Engine.GET("/chat", m.Chat)
	//fmt.Println("调用完chat")
	Engine.GET("/chat/createcom.shtml", m.createcom)

}
func (m *ChatController) createcom(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "/chat/createcom.shtml", nil)
}

func (m *ChatController) Chat(ctx *gin.Context) {
	id := ctx.Query("id")
	token := ctx.Query("token")
	userId, _ := strconv.ParseInt(id, 10, 64)

	isValid := checkToken(userId, token)

	if !isValid {
		tool.Fail(ctx, "不匹配的token")
		return
	}
	//准备将此http请求升级成websocket请求
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isValid
		},
	}
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		tool.Fail(ctx, fmt.Sprintln(err))
	}

	var node *Node

	//查看是不是老用户
	if CheckExist(userId) {
		//此为老用户回归的步骤
		*node = userBack(userId)
		node.Conn = conn
		node.Status = true
		node.GroupSets = set.New(set.ThreadSafe)
	} else {
		node = &Node{
			Conn:      conn, //外部调用此路由时的conn
			DataQueue: make(chan []byte, 50),
			//用于群聊，如果选用根据群id后以遍历用户id的方式向群中的所有户发信息的方法，那么就不需要这一个字段
			GroupSets: set.New(set.ThreadSafe),
			Status:    true,
		}
	}

	//初始化时，获取接入用户的全部群id
	contact := service.ContactService{}
	comIds := contact.SearchComunityIds(userId)
	//将数据库中拿取出来的值刷新node的groupset中
	for _, comId := range comIds {
		node.GroupSets.Add(comId)
	}
	//绑定userid和node。recvproc的dispatch函数的data参数反序列化后可以得到其中的Dstid即目标对象id，我们可以根据目标对象id拿到目标对应的conn连接，然后向它的队列里发送数据，由于是并发，其他协程不停处理向目标对象的sendproc.
	rwlocker.Lock()
	clientMap[userId] = node
	rwlocker.Unlock()
	//发送消息到节点结构体的数据通道中
	//sendMsg(userId, []byte("hello,world!"))
	//将节点结构体的数据通道的数据发送给节点的conn
	fmt.Println("start********当前工作协程数目", runtime.NumGoroutine())
	//接收处理
	go recvproc(node)
	go sendproc(userId, node)
	fmt.Println("END********当前工作协程数目", runtime.NumGoroutine())
}

//流程应该是后端先被触发接收函数recvproc接收到data(由前端传过来，其中会包含着目标dstId),然后执行dispatch函数根据不同的文件类型执行流程。其中有一个流程就是执行sendmsg函数即向目标对象的数据队列发送数据。接着并发执行着的sendproc,会从node的队列中拿去数据发送给目标websocket对象
//发送协程
func sendproc(userId int64, node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			fmt.Printf("\n要发送给node对应conn的值%v", string(data))
			if err != nil {
				fmt.Println("sendproc发送消息失败", err)
				return
			}
			//TODO 向历史消息表存入消息
			storeHistoryMessage(userId, data)
		}
	}
}

//发送协程
func recvproc(node *Node) {
	for {
		data := make([]byte, 0)
		ty, data, err := node.Conn.ReadMessage()
		if ty == 1000 {
			node.Status = false
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(data)
		//broadMsg(data)
		fmt.Printf("%s\n", data)
	}
}

var udpsendchan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendchan <- data
}

func udpsendproc() {
	//使用udp拨号
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(192, 168, 0, 255),
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	//通过得到的con发送消息
	for {
		select {
		case data := <-udpsendchan:
			_, err := con.Write(data)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

func udprecvproc() {
	//监听udp广播端口
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer con.Close()
	for {
		//每次只处理512字节
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		dispatch(buf[0:n])
	}
}

func dispatch(data []byte) {
	//解析data
	msg := model.Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	//根据cmd消息类型处理
	switch msg.Cmd {
	case model.CMD_SINGLE_MSG:
		sendMsg(msg.Dstid, data)
		//群聊消息
	case model.CMD_ROOM_MSG:
		for k, node := range clientMap {
			if node.GroupSets.Has(msg.Dstid) {
				fmt.Printf("\n当前节点的Id:%v,消息发送人的Id:%v,消息目的(群组)的Id:%v\n", k, msg.Userid, msg.Dstid)
				node.DataQueue <- data
			}
		}
	case model.CMD_HEART:
		// 心跳为保证网络持久性。数据来源若是心跳事件一般什么都不做,更新时间戳
		clientMap[msg.Userid].LastHeartbeat = time.Now().UnixNano()

	}
}

func AddGroupId(userId, gId int64) {
	//map涉及到可能并发的时候，使用map先上读写锁
	rwlocker.Lock()
	node, ok := clientMap[userId]
	if ok {
		node.GroupSets.Add(gId)
	}
	//解锁
	rwlocker.Unlock()

}

//发送消息，先放到该节点的数据通道当中(考虑到并发所以有通道)

func sendMsg(userId int64, msg []byte) {
	rwlocker.Lock()
	node, ok := clientMap[userId]
	rwlocker.Unlock()
	if ok {
		node.DataQueue <- msg
	} else {
		if CheckExist(userId) {
			//如果存在先将离线消息写入redis，待对方登录时从中读取
			err := tool.CacheStoreMsg(msg, userId)
			if err != nil {
				return
			}
		}
	}
}

func CheckExist(userId int64) bool {
	db := tool.XormMysqlEngine
	clientData := model.User{}
	n, err := db.Table("user_stored").Where("userid = ?", userId).Get(&clientData)
	if err != nil || n == false {
		log.Fatal(err)
		return false
	}

	return true

}

func checkToken(userId int64, token string) bool {
	//从数据库里查询并操作
	srv := service.UserService{}
	user := srv.Find(userId)
	return user.Token == token
}

//清理超时连接
func cleanConnection() {
	rwlocker.RLock()
	defer rwlocker.RUnlock()
	curTime := time.Now().UnixNano()
	for userid, node := range clientMap {
		if node.LastHeartbeat+int64(maxIdleConnectionTime) > curTime {
			fmt.Println("心跳超时。。断开连接")
			//原子修改节点状态值
			pointer := unsafe.Pointer(&node.Status)
			newBool := false
			atomic.StorePointer(&pointer, unsafe.Pointer(&newBool))
			//node.Status = false
			//TODO 将收件箱里的信息存储至redis离线消息表及mysql的历史消息表中
			procOfflineMessage(userid, node)
			node.Conn.Close()
		}
	}
}

//定期清理本地缓存表
func cleanClientMap() {
	rwlocker.RLock()
	curTime := time.Now().UnixNano()
	for key, node := range clientMap {
		if node.LastHeartbeat+int64(maxIdleduration) > curTime && node.Status == false {
			fmt.Println("用户长时间未登录。。清理连接")
			rwlocker.RUnlock()
			rwlocker.Lock()
			//删除对应用户
			delete(clientMap, key)
			rwlocker.Unlock()
			return
		}
	}
	rwlocker.RUnlock()
}

//存储线下信息
func procOfflineMessage(userid int64, node *Node) {
	if node.DataQueue == nil {
		return
	}
	for {
		//每次存储一条信息，并设置过期时间7天
		bytes, ok := <-node.DataQueue
		if ok == true {
			//存储到redis中
			err := tool.CacheStoreMsg(bytes, userid)
			if err != nil {
				return
			}
			//存入历史消息中
			//TODO 尝试使用消息队列将历史消息写入数据库，解耦流程
			storeHistoryMessage(userid, []byte(tool.BaseEncode(bytes)))
		} else {
			return
		}
	}
}

//存储历史消息
func storeHistoryMessage(userid int64, data []byte) {
	db := tool.XormMysqlEngine
	msg := model.HistoryMessage{}
	err := json.Unmarshal(data, &msg)
	//当前unix时间戳，纳秒
	msg.Timestamp = time.Now().UnixNano()
	if err != nil {
		return
	}
	index := strconv.FormatInt(userid, 10)
	index += strconv.FormatInt(msg.Timestamp, 10)
	_, err = db.Table("history_table").Where("userid = ?", index).Insert(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func userBack(userId int64) Node {
	//查看是不是老用户
	//此为老用户回归的步骤
	node := Node{
		Conn:          nil,
		DataQueue:     make(chan []byte, 50),
		GroupSets:     nil,
		LastHeartbeat: time.Now().UnixNano(),
		Status:        false,
	}
	db := tool.XormMysqlEngine
	nodeData := make([]struct {
		Data []byte
	}, 0)

	//分别获取Node的群信息和data信息
	db.Table("archive_user_file").Join("Inner", "user_data_queue", "archive_user_file.id = user_data_queue.userid").Cols("user_data_queue.data").OrderBy("user_data_queue.timestamp").Find(&nodeData)

	//-----------------
	//可以直接通过群表获得，不需要这样
	//nodeGroup := struct {
	//	BinarySet []byte
	//}{}
	//db.Table("archive_user_file").Cols("archive_user_file.groupsets").Get(&nodeGroup)
	////将群信息存入node的GroupSets
	//node.GroupSets = tool.GobDecode(nodeGroup.BinarySet, node.GroupSets).(set.Interface)
	//---------

	//将数据存入node的DataQueue
	for _, s := range nodeData {
		node.DataQueue <- s.Data
	}
	return node
}

//断线检测,检测节点断线
func CheckUserStatus() {
	for _, node := range clientMap {
		//判断是否已经断开连接
		if node.Conn == nil {
			node.Status = false
		}
	}
}

func init() {
	go cleanClientMap()
	go cleanConnection()
	go CheckUserStatus()
	//当前未使用
	go udprecvproc()
	go udpsendproc()
}

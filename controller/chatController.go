package controller

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"net"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"websocketIMProject/model"
	"websocketIMProject/service"
	"websocketIMProject/tool"
)

type Node struct {
	Conn *websocket.Conn
	//并行转串行,
	DataQueue chan []byte
	GroupSets set.Interface
}

var clientMap map[int64]*Node = make(map[int64]*Node, 0)

var rwlocker sync.RWMutex

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
	node := &Node{
		Conn:      conn, //外部调用此路由时的conn
		DataQueue: make(chan []byte, 50),
		//用于群聊，如果选用根据群id后以遍历用户id的方式向群中的所有户发信息的方法，那么就不需要这一个字段
		GroupSets: set.New(set.ThreadSafe),
	}
	//初始化时，获取接入用户的全部群id
	contact := service.ContactService{}
	comIds := contact.SearchComunityIds(userId)
	//将数据库中拿取出来的值刷新node的的groupset中
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
	go sendproc(node)
	fmt.Println("END********当前工作协程数目", runtime.NumGoroutine())
}

//流程应该是后端先被触发接收函数recvproc接收到data(由前端传过来，其中会包含着目标dstId),然后执行dispatch函数根据不同的文件类型执行流程。其中有一个流程就是执行sendmsg函数即向目标对象的数据队列发送数据。接着并发执行着的sendproc,会从node的队列中拿去数据发送给目标websocket对象
//发送协程
func sendproc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			fmt.Printf("\n要发送给node对应conn的值%v", string(data))
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}
}

//发送协程
func recvproc(node *Node) {
	for {
		data := make([]byte, 0)
		_, data, err := node.Conn.ReadMessage()
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
		// 心跳为保证网络持久性。数据来源若是心跳事件一般什么都不做

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
	}
}

func checkToken(userId int64, token string) bool {
	//从数据库里查询并操作
	srv := service.UserService{}
	user := srv.Find(userId)
	return user.Token == token
}

func init() {
	go udprecvproc()
	go udpsendproc()
}

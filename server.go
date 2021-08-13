package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip string
	Port int
	OnlineMap map[string]*User
	OnlineMapLock sync.RWMutex

	// 消息广播的 chan
	MessageChan chan string
}

// NewServer 创建一个 listener server
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
		MessageChan: make(chan string),
		OnlineMap: make(map[string]*User),
	}
	return server
}

// Start 开始运行监听
func (s *Server) Start(){
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil{
		fmt.Println("net listen error: ", err)
		return
	}
	defer listener.Close()
	// 开启监听广播消息的协程
	go s.ListenMessage()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listen accept err :", err)
			continue
		}
		// 开启针对此连接的协程
		go s.Handler(conn)
	}
}

// BroadCat 广播消息
func (s *Server) BroadCat(user *User, msg string){
	s.MessageChan <- msg
}

// Handler 用户的goroutine
func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)
	user.OnLine()
	// 监听用户输入
	for{
		msg := make([]byte, 4096)
		lenMsg, err := conn.Read(msg)
		if lenMsg == 0{
			// 下线
			user.OffLine()
			return
		}
		if err != nil && err != io.EOF{
			fmt.Println("读取消息失败: ", err)
			return
		}
		user.DoMessage(string(msg[:lenMsg-1]))
	}
}

// ListenMessage 监听广播通道 如果有消息就发送给全部用户
func (s *Server) ListenMessage(){
	for {
		message := <- s.MessageChan
		s.OnlineMapLock.Lock()
		for _, user := range s.OnlineMap {
			user.C <- message
		}
		s.OnlineMapLock.Unlock()
	}
}

// AddUser 新增在线用户
func (s *Server) AddUser(user *User) bool {
	s.OnlineMapLock.Lock()
	s.OnlineMap[user.Address] = user
	s.OnlineMapLock.Unlock()
	return true
}

// DelUser 删除下线用户
func (s *Server) DelUser(user *User) bool{
	s.OnlineMapLock.Lock()
	delete(s.OnlineMap, user.Address)
	s.OnlineMapLock.Unlock()

	defer user.Conn.Close()
	return true
}

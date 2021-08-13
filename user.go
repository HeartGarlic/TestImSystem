package main

import "net"

type User struct {
	Name    string
	Address string
	Conn    net.Conn
	C       chan string
	Server  *Server
}

// NewUser 新增一个用户
func NewUser(conn net.Conn, service *Server) *User {
	user := &User{
		Name:    conn.RemoteAddr().String(),
		Address: conn.RemoteAddr().String(),
		Conn:    conn,
		C:       make(chan string),
		Server:  service,
	}

	// 新建用户的时候就开始监听是否有消息需要发送给用户
	go user.ListenMessage()

	return user
}

// ListenMessage 监听消息如果有消息就发送给用户
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.SendMessage(msg)
	}
}

// SendMessage 给用户发送消息
func (u *User) SendMessage(msg string) {
	u.Conn.Write([]byte(msg + "\n"))
}

// DoMessage 用户处理消息
func (u *User) DoMessage(msg string) {
	// 增加更改用户名的功能
	if len(msg) > 7 && msg[:7] == "rename|"{
		u.Name = msg[7:]
		u.SendMessage("您已经更改用户名为:" + u.Name)
	}else if len(msg) == 3 && msg=="who" {
		u.Server.OnlineMapLock.Lock()
		for _, onlineUser := range u.Server.OnlineMap{
			u.SendMessage(onlineUser.Name)
		}
		u.Server.OnlineMapLock.Unlock()
	}else{
		u.Server.BroadCat(u, "["+ u.Name +"]say:" + msg)
	}
}

// OffLine 用户下线处理
func (u *User) OffLine(){
	u.Server.DelUser(u)
	u.Server.BroadCat(u, "["+u.Address+"]" + "已下线")
}

// OnLine 用户上线处理
func (u *User) OnLine() {
	u.Server.AddUser(u)
	u.Server.BroadCat(u, "["+u.Address+"]" + "已上线")
}

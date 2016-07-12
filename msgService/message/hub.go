package message

import (
	"fmt"

	"github.com/longchat/longChat-Server/common/protoc"
	"golang.org/x/net/context"
)

var groupUser map[int64]map[int64]struct{}
var userConn map[int64][]*Conn
var command chan interface{}

var msgChan chan *protoc.MessageReq

type connAdd struct {
	conn    *Conn
	userId  int64
	groupId []int64
}

type connDel struct {
	conn    *Conn
	userId  int64
	groupId []int64
}

func controller() {
	connId := 0
	userConn = make(map[int64][]*Conn, 500)
	groupUser = make(map[int64]map[int64]struct{}, 100)
	command = make(chan interface{})
	msgChan = make(chan *protoc.MessageReq, 800)
	for {
		select {
		case item := <-command:
			switch value := item.(type) {
			case connAdd:
				addConn(value, &connId)
			case connDel:
				removeConn(value, &connId)
			}
		case message := <-msgChan:
			processMessage(message)
		}
	}
}

func processMessage(msg *protoc.MessageReq) {
	if msg.GroupId > 0 {
		users, isok := groupUser[msg.GroupId]
		if isok {
			for k, _ := range users {
				conns := userConn[k]
				for i := range conns {
					conns[i].send <- msg
				}
			}
		}
	}
}

func removeConn(c connDel, id *int) {
	(*id)--
	user, _ := userConn[c.userId]
	var i int
	length := len(user)
	for i = range user {
		if user[i].id == c.conn.id {
			break
		}
	}

	if length == 1 {
		delete(userConn, c.userId)
		for _, data := range c.groupId {
			group, _ := groupUser[data]
			delete(group, c.userId)
		}
	} else {
		if i < (length - 1) {
			copy(user[i:length-1], user[i+1:length])
		}
		user = user[:length-1]
		userConn[c.userId] = user
	}

}

func addConn(c connAdd, id *int) {
	(*id)++
	c.conn.id = *id
	user, isok := userConn[c.userId]
	if !isok {
		user = make([]*Conn, 1)
		user[0] = c.conn
	} else {
		user = append(user, c.conn)
	}
	userConn[c.userId] = user
	if !isok {
		for _, data := range c.groupId {
			group, isok := groupUser[data]
			if !isok {
				group = make(map[int64]struct{}, 20)
				reply, err := router.GroupUp(context.Background(), &protoc.GroupUpReq{GroupId: data, ServerAddr: msgAddr})
				if err != nil || reply.StatusCode != 0 {
					fmt.Printf("group up to router failed!err:=%v,err:=%v", err, reply.Err)
				}
			}
			var a struct{}
			group[c.userId] = a
			groupUser[data] = group
		}
	}
}

var storeChan chan *protoc.MessageReq
var readChan chan userRead

type userRead struct {
	userId    int64
	groupRead map[int64]int64
}

func msgPersist() {
	userReadMap := make(map[int64]map[int64]int64, 200)
	for {
		select {
		case read := <-readChan:
			markRead(userReadMap, read)
		}
	}
}

func markRead(readMap map[int64]map[int64]int64, read userRead) {

}

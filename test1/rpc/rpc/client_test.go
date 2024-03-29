package rpc

import (
	"encoding/gob"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// 定义用户对象
type User struct {
	Name string
	Age  int
}

// 用于测试用户查询的方法
func queryUser(uid int) (User, error) {
	user := make(map[int]User)
	// 假数据
	user[0] = User{"zs", 20}
	user[1] = User{"ls", 21}
	user[2] = User{"ww", 22}
	// 模拟查询用户
	if u, ok := user[uid]; ok {
		return u, nil
	}
	return User{}, fmt.Errorf("%d err", uid)
}

func TestRPC(t *testing.T) {
	//该名称将标识作为接口变量发送或接收的值的具体类型。只有将作为接口值的实现传输的类型才需要注册
	gob.Register(User{})
	addr := ":8080"
	srv := NewServer(addr)
	//注册服务器方法
	srv.Register("query", queryUser)
	go srv.Run()
	//客户端获取链接
	time.Sleep(time.Second)
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)
	cli := NewClient(conn)
	var query func(uid int) (User, error)
	errchan := cli.CallRPC("query", &query)
	u, err := query(1)
	linkErr := <-errchan
	require.NoError(t, linkErr)
	require.NoError(t, err)
	require.Equal(t, u, User{"ls", 21})
}

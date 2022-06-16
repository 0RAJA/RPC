package rpc

import (
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSession_ReadWriter(t *testing.T) {
	// 定义地址
	addr := "127.0.0.1:8000"
	my_data := "hello"
	// 等待组定义
	wg := sync.WaitGroup{}
	wg.Add(2)
	ok := make(chan struct{})
	// 写数据的协程
	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", addr)
		require.NoError(t, err)
		ok <- struct{}{}
		conn, _ := lis.Accept()
		s := NewSession(conn)
		err = s.Write([]byte(my_data))
		require.NoError(t, err)
	}()

	// 读数据的协程
	go func() {
		<-ok
		defer wg.Done()
		conn, err := net.Dial("tcp", addr)
		require.NoError(t, err)
		s := NewSession(conn)
		data, err := s.Read()
		require.NoError(t, err)
		// 最后一层校验
		require.Equal(t, data, []byte(my_data))
	}()
	wg.Wait()
}

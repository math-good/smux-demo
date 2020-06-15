package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xtaci/smux"
	"math/rand"
	"net"
	"runtime"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
	if runtime.GOOS == "windows" {
		//windows机器设置为debug级别
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

/**
一个生成随机数的tcp服务
客户端发送'R', 'A', 'N', 'D'，服务返回一个随机数
*/
func main() {
	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	log.Info().Msg("随机数服务启动，监听9000端口")
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		go SessionHandler(conn)
	}
}

/**
处理会话
每个tcp连接生成一个会话session
*/
func SessionHandler(conn net.Conn) {
	session, err := smux.Server(conn, nil)
	if err != nil {
		panic(err)
	}
	log.Info().Msgf("收到客户端连接，创建新会话，对端地址：%s", session.RemoteAddr().String())

	for !session.IsClosed() {
		stream, err := session.AcceptStream()
		if err != nil {
			fmt.Println(err.Error())
			break
		}
		go StreamHandler(stream)
	}
	log.Info().Msgf("客户端连接断开，销毁会话，对端地址：%s", session.RemoteAddr().String())
}

/**
流数据处理
*/
func StreamHandler(stream *smux.Stream) {
	buffer := make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		log.Error().Msgf("流id：%d，异常信息：%s", stream.ID(), err.Error())
		stream.Close()
		return
	}
	cmd := buffer[:n]
	if bytes.Equal(cmd, []byte{'R', 'A', 'N', 'D'}) {
		rand := rand.Uint64()
		response := make([]byte, 8)
		binary.BigEndian.PutUint64(response, rand)
		stream.Write(response)
		log.Debug().Msgf("收到客户端数据，流id：%d，随机数：%d， 响应数据：%v", stream.ID(), rand, response)
	} else {
		log.Warn().Msgf("收到未知请求命令，流id：%d，请求命令：%v", stream.ID(), cmd)
	}
}

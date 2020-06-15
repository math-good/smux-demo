package main

import (
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xtaci/smux"
	"net"
	"net/http"
	"runtime"
)

/**
随机数服务客户端连接
*/
var randClient *smux.Session

func init() {
	if runtime.GOOS == "windows" {
		//windows机器设置为debug级别
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
	//连接后端随机数服务
	conn, err := net.Dial("tcp", ":9000")
	if err != nil {
		log.Warn().Msg("随机数服务未启动")
		panic(err)
	}
	session, err := smux.Client(conn, nil)
	if err != nil {
		log.Error().Msg("打开会话失败")
		panic(err)
	}
	randClient = session
}

/**
一个api网关，对外提供api接口
调用随机数服务来获取随机数
*/
func main() {
	defer randClient.Close()
	http.HandleFunc("/rand", RandHandler)
	http.ListenAndServe(":8080", nil)
}

/**
随机数接口
*/
func RandHandler(w http.ResponseWriter, r *http.Request) {
	stream, err := randClient.OpenStream()
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
	} else {
		log.Debug().Msgf("收到请求，打开流成功，流id：%d", stream.ID())
		defer stream.Close()
		stream.Write([]byte{'R', 'A', 'N', 'D'})
		buffer := make([]byte, 1024)
		n, err := stream.Read(buffer)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, err.Error())
		} else {
			response := buffer[:n]
			var rand = binary.BigEndian.Uint64(response)
			log.Debug().Msgf("收到服务端数据，流id：%d，随机数：%d， 响应数据：%v", stream.ID(), rand, response)
			fmt.Fprintf(w, "%d", rand)
		}
	}
}

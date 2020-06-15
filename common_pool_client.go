package main

import (
	"context"
	"encoding/binary"
	"fmt"
	cpool "github.com/jolestar/go-commons-pool/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/xtaci/smux"
	"net"
	"net/http"
	"runtime"
)

var commonPool *cpool.ObjectPool
var ctx = context.Background()

func init() {
	if runtime.GOOS == "windows" {
		//windows机器设置为debug级别
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	factory := cpool.NewPooledObjectFactorySimple(NewSessionCpool)

	commonPool = cpool.NewObjectPoolWithDefaultConfig(ctx, factory)
	commonPool.Config.MaxTotal = 20
}

/**
连接池生成新会话函数
*/
func NewSessionCpool(ctx context.Context) (interface{}, error) {
	log.Debug().Msg("连接池中生成一个连接")
	//连接后端随机数服务
	conn, err := net.Dial("tcp", ":9000")
	if err != nil {
		log.Warn().Msg("随机数服务未启动")
		panic(err)
	}
	//随机数服务客户端连接
	session, err := smux.Client(conn, nil)
	if err != nil {
		log.Error().Msg("打开会话失败")
		panic(err)
	}
	return session, err
}

/**
一个api网关，对外提供api接口
调用随机数服务来获取随机数

通过sync.Pool实现“连接池” ！！！ 不推荐这种方式，sync.Pool的种种特性不适合作为连接池
*/
func main() {
	http.HandleFunc("/rand", CommonPoolRandHandler)
	http.ListenAndServe(":8080", nil)
}

/**
随机数接口
*/
func CommonPoolRandHandler(w http.ResponseWriter, r *http.Request) {
	obj, err := commonPool.BorrowObject(ctx)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, err.Error())
		return
	}
	client := obj.(*smux.Session)
	stream, err := client.OpenStream()
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
	commonPool.ReturnObject(ctx, obj)
}

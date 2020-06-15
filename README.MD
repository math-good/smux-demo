##smux-demo

基于github.com/xtaci/smux的连接多路复用

使用github.com/jolestar/go-commons-pool/v2作为连接池

server.go为随机数服务，客户端发送'R', 'A', 'N', 'D'，服务返回一个随机数

client有三种，simple_client，pool_client和common_pool_client

client充当与一个apiservice角色，对外提供api接口，对内连接randservice获取随机数，如下图

```text
               n link                  10 link
+-----------+          +-------------+           +---------------+
|           <---------->             <----------->               |
|  browser  <---------->  apiservice <----------->  randservice  |
|           <---------->             <----------->               |
+-----------+          +-------------+           +---------------+
```



simple_client：简单客户端

pool_client：使用sync.Pool作为连接池（不建议使用，sync.Pool没有最大连接数限制，连接释放不通知）

common_pool_client：使用go-commons-pool作为连接池

通过连接池来保证apiservice和randservice之前的连接个数固定，不受browser个数影响

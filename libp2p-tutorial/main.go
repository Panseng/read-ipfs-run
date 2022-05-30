package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/libp2p/go-libp2p"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	multiaddr "github.com/multiformats/go-multiaddr"
)

func main() {
	node, err := libp2p.New(
		// libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/2000"),
		libp2p.ListenAddrStrings("/ip4/127.0.0.1/tcp/0"), // 配置随机端口而不是硬编码端口
		libp2p.Ping(false),
	)
	if err != nil {
		panic(err)
	}
	// fmt.Println("libp2p listening addresses: ", node.Addrs())

	// 配置个性化的 ping 协议
	pingService := &ping.PingService{Host: node}
	node.SetStreamHandler(ping.ID, pingService.PingHandler)

	// 配置 p2p 节点
	peerInfo := peerstore.AddrInfo{
		ID:    node.ID(),
		Addrs: node.Addrs(),
	}
	addrs, err := peerstore.AddrInfoToP2pAddrs(&peerInfo)
	if err != nil{
		panic(err)
	}
	fmt.Println("libp2p node address: ", addrs[0])

	if len(os.Args) > 1 {
		addr, err := multiaddr.NewMultiaddr(os.Args[1])
		if err != nil {
			panic(err)
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			panic(err)
		}
		if err := node.Connect(context.Background(), *peer); err != nil {
			panic(err)
		}
		fmt.Println("sending 5 ping msg to ", addr)
		ch := pingService.Ping(context.Background(), peer.ID)
		for i := 0; i < 5; i++ {
			res := <-ch
			fmt.Println("got ping response, RRT: ", res.RTT)
		}
		// res := <-ch
		// fmt.Println("ping response: ", res)
	} else {
		// 以下代码，会使得主进程只有收到 ctrl-c 才会执行后面的 close代码
		// 相当于是被阻塞
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
	}

	// 关闭节点
	if err = node.Close(); err != nil {
		panic(err)
	}
}

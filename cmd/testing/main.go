package main

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
)

func forward(name string, source, target *net.TCPConn, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Infof("forward start %s", name)
	_, err := io.Copy(target, source)
	log.Infof("forward end %s %v", name, err)
	if err == nil {
		// 如果源关闭了写端，关闭目标的写端
		target.CloseWrite()
	} else if err != nil {
		// handle other errors
	}
}

func handleConnection(clientConn, serverConn *net.TCPConn) {
	var wg sync.WaitGroup
	wg.Add(2)
	log.Info("forward start")

	// 从客户端到服务器
	go forward("L->R", clientConn, serverConn, &wg)

	// 从服务器到客户端
	go forward("L<-R", serverConn, clientConn, &wg)

	wg.Wait()
	log.Info("forward end")
}

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}

	for {
		client, err := ln.Accept()
		if err != nil {
			// handle error
		}

		server, err := net.Dial("tcp", "127.0.0.1:8080") // 替换为您的目标服务器地址和端口
		if err != nil {
			// handle error
		}

		go handleConnection(client.(*net.TCPConn), server.(*net.TCPConn))
	}
}

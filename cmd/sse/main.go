package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// Message 定义SSE消息结构
type Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Broker 用于管理所有客户端连接
type Broker struct {
	// 客户端连接通道
	clients map[chan Message]bool
	// 用于同步的锁
	mutex sync.Mutex
}

// NewBroker 创建新的Broker实例
func NewBroker() *Broker {
	return &Broker{
		clients: make(map[chan Message]bool),
	}
}

// AddClient 添加新客户端
func (b *Broker) AddClient(client chan Message) {
	b.mutex.Lock()
	b.clients[client] = true
	b.mutex.Unlock()
	log.Printf("Client added. %d registered clients", len(b.clients))
}

// RemoveClient 移除客户端
func (b *Broker) RemoveClient(client chan Message) {
	b.mutex.Lock()
	delete(b.clients, client)
	b.mutex.Unlock()
	log.Printf("Client removed. %d registered clients", len(b.clients))
}

// Broadcast 向所有客户端广播消息
func (b *Broker) Broadcast(msg Message) {
	b.mutex.Lock()
	for client := range b.clients {
		client <- msg
	}
	b.mutex.Unlock()
}

func main() {
	broker := NewBroker()

	// 启动一个goroutine模拟数据变化
	go func() {
		for {
			// 模拟数据变化
			broker.Broadcast(Message{
				Event: "time",
				Data:  time.Now().Format("2006-01-02 15:04:05"),
			})

			broker.Broadcast(Message{
				Event: "stats",
				Data:  fmt.Sprintf("Active clients: %d", len(broker.clients)),
			})

			time.Sleep(1 * time.Second)
		}
	}()

	// 设置路由
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		sseHandler(w, r, broker)
	})
	http.HandleFunc("/", homeHandler)

	// 启动服务器
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>SSE Test</title>
	</head>
	<body>
		<h1>Server-Sent Events Demo</h1>
		<div id="time"></div>
		<div id="stats"></div>
		<script>
			const eventSource = new EventSource('/events');
			
			eventSource.addEventListener('time', function(e) {
				document.getElementById('time').innerHTML = 'Time: ' + e.data;
			});
			
			eventSource.addEventListener('stats', function(e) {
				document.getElementById('stats').innerHTML = 'Stats: ' + e.data;
			});
		</script>
	</body>
	</html>
	`)
}

func sseHandler(w http.ResponseWriter, r *http.Request, broker *Broker) {
	// 设置SSE所需的响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 创建一个通道用于当前客户端
	messageChan := make(chan Message)

	// 注册客户端
	broker.AddClient(messageChan)

	// 确保客户端退出时注销
	defer func() {
		broker.RemoveClient(messageChan)
		close(messageChan)
	}()

	// 监听连接关闭
	notify := r.Context().Done()

	// 保持连接打开
	for {
		select {
		case <-notify:
			// 客户端断开连接
			return
		case msg := <-messageChan:
			// 将消息编码为JSON
			jsonData, err := json.Marshal(msg.Data)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			// 发送SSE格式的消息
			fmt.Fprintf(w, "event: %s\n", msg.Event)
			fmt.Fprintf(w, "data: %s\n\n", jsonData)
			w.(http.Flusher).Flush()
		}
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v3"
)

func main() {

	var offerChannel = make(chan string)

	var answerChannel = make(chan string)

	var dataChannel = make(chan string)

	var offerSDP string

	go setupWebRTC(offerChannel, answerChannel, dataChannel)

	offerSDP = <-offerChannel

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.GET("/offer", func(c *gin.Context) {

		c.String(http.StatusOK, offerSDP)
	})
	r.POST("/answer", func(c *gin.Context) {
		answerSdp := c.PostForm("answer")
		answerChannel <- answerSdp
		c.String(http.StatusOK, "Answer SDP received")
	})
	r.POST("/data", func(c *gin.Context) {
		data := c.PostForm("data")
		dataChannel <- data
		c.String(http.StatusOK, "Data received")
	})
	r.Run()

}
func setupWebRTC(offerChannel chan string, answerChannel chan string, dataChn chan string) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// 创建 PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		fmt.Printf("创建 PeerConnection 失败: %v\n", err)
		return
	}
	defer peerConnection.Close()

	// 创建数据通道
	dataChannel, err := peerConnection.CreateDataChannel("data", nil)
	if err != nil {
		fmt.Printf("创建数据通道失败: %v\n", err)
		return
	}

	// 处理数据通道状态
	dataChannel.OnOpen(func() {
		fmt.Println("数据通道已打开")
		go handleInput(peerConnection, dataChannel, dataChn)
	})

	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Printf("收到消息: %s\n", string(msg.Data))
	})

	// 创建 Offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		fmt.Printf("创建 Offer 失败: %v\n", err)
		return
	}
	// 将 Offer SDP 发送到 channel
	offerSDP, err := json.Marshal(offer)
	if err != nil {
		fmt.Printf("序列化 Offer SDP 失败: %v\n", err)
		return
	}
	offerChannel <- string(offerSDP)
	fmt.Println("Offer SDP 已发送到 channel")

	// 设置本地描述
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		fmt.Printf("设置本地描述失败: %v\n", err)
		return
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete
	go func() {

		// 等待 ICE 候选收集完成

		// 等待 Answer SDP
		answerSDP := <-answerChannel
		fmt.Println("收到 Answer SDP")
		// 解析 Answer SDP
		if answerSDP == "" {
			fmt.Println("没有收到有效的 Answer SDP")
			return
		}

		var answer webrtc.SessionDescription
		if err := json.Unmarshal([]byte(answerSDP), &answer); err != nil {
			fmt.Printf("解析 Answer SDP 失败: %v\n", err)
			return
		}

		// 设置远程描述
		if err = peerConnection.SetRemoteDescription(answer); err != nil {
			fmt.Printf("设置远程描述失败: %v\n", err)
			return
		}
	}()

}

// 处理用户输入（文本或文件）
func handleInput(_ *webrtc.PeerConnection, dc *webrtc.DataChannel, data chan string) {

	for {
		select {
		case input := <-data:
			dc.SendText(input)
		default:

		}

	}
}

// 发送文件
func sendFile(dc *webrtc.DataChannel, filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("打开文件失败: %v\n", err)
		return
	}
	defer file.Close()

	// 发送文件名
	filename := filepath.Base(filePath)
	if err := dc.SendText("FILE:" + filename); err != nil {
		fmt.Printf("发送文件名失败: %v\n", err)
		return
	}

	// 分片发送文件内容
	buffer := make([]byte, 16384) // WebRTC 数据通道的最大消息大小
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Printf("读取文件失败: %v\n", err)
			return
		}
		if n == 0 {
			break
		}

		if err := dc.Send(buffer[:n]); err != nil {
			fmt.Printf("发送文件数据失败: %v\n", err)
			return
		}
		time.Sleep(10 * time.Millisecond) // 防止数据通道阻塞
		fmt.Println("发送数据片段:", n, "字节")
	}

	fmt.Printf("文件 %s 已发送\n", filename)
}

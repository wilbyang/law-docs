package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
)

func main() {
	// 创建 WebRTC 配置
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

	var currentFile *os.File
	var currentFilename string

	// 监听数据通道
	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		fmt.Println("数据通道已连接")
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			if msg.IsString {
				message := string(msg.Data)
				if strings.HasPrefix(message, "FILE:") {
					// 接收到文件开始
					currentFilename = strings.TrimPrefix(message, "FILE:") + ".txt"
					currentFile, err = os.Create(currentFilename)
					if err != nil {
						fmt.Printf("创建文件 %s 失败: %v\n", currentFilename, err)
						return
					}
					fmt.Printf("开始接收文件: %s\n", currentFilename)
				} else {
					// 接收到文本消息
					fmt.Printf("收到消息: %s\n", message)
				}
			} else {
				fmt.Printf("收到二进制数据，长度: %d\n", len(msg.Data))
				fmt.Printf("接收到 bytes: %v", msg.Data)
				// 接收到文件数据
				if currentFile != nil {
					_, err := currentFile.Write(msg.Data)
					if err != nil {
						fmt.Printf("写入文件 %s 失败: %v\n", currentFilename, err)
						return
					}
				}
			}
		})

		dc.OnClose(func() {
			if currentFile != nil {
				currentFile.Close()
				fmt.Printf("文件 %s 接收完成\n", currentFilename)
				currentFile = nil
				currentFilename = ""
			}
		})
	})

	// 读取文件内容
	offerBytes, err := os.ReadFile("offer.sdp")
	if err != nil {
		fmt.Printf("读取 Offer SDP 文件失败: %v\n", err)
		return
	}

	var offer webrtc.SessionDescription
	if err := json.Unmarshal(offerBytes, &offer); err != nil {
		fmt.Printf("解析 Offer SDP 失败: %v\n", err)
		return
	}

	// 设置远程描述
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		fmt.Printf("设置远程描述失败: %v\n", err)
		return
	}

	// 创建 Answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		fmt.Printf("创建 Answer 失败: %v\n", err)
		return
	}

	// 设置本地描述
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		fmt.Printf("设置本地描述失败: %v\n", err)
		return
	}

	// 等待 ICE 候选收集完成
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete

	// 输出 Answer SDP
	answerSDP, err := json.Marshal(peerConnection.LocalDescription())
	if err != nil {
		fmt.Printf("序列化 Answer SDP 失败: %v\n", err)
		return
	}

	// 写入 Answer SDP 到文件
	if err := os.WriteFile("answer.sdp", answerSDP, 0644); err != nil {
		fmt.Printf("写入 Answer SDP 文件失败: %v\n", err)
		return
	}
	fmt.Println("Answer SDP 已保存到 answer.sdp")

	// 保持程序运行
	select {}
}

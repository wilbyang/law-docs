package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6380"})

	for {
		res, err := client.XRead(ctx, &redis.XReadArgs{
			Streams: []string{"alerts", "$"}, // "$" 表示从最新位置开始
			Count:   1,
			Block:   0, // 阻塞等待
		}).Result()
		if err != nil {
			log.Fatalf("读取事件失败: %v", err)
		}

		for _, stream := range res {
			for _, msg := range stream.Messages {
				fmt.Printf("处理事件: %v\n", msg.Values)
			}
		}
	}
}

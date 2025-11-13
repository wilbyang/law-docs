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

	event := map[string]interface{}{"message": "Critical alert! Server down."}

	_, err := client.XAdd(ctx, &redis.XAddArgs{
		Stream: "alerts",
		Values: event,
	}).Result()
	if err != nil {
		log.Fatalf("发布事件失败: %v", err)
	}
	fmt.Println("事件发布成功")
}

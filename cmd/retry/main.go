package main

import (
	"errors"
	"fmt"
	"time"
)

// Retryable 定义可重试操作的接口
type Retryable interface {
	Attempt() error
	ShouldRetry(error) bool
	Backoff(attempt int) time.Duration
}

// Retry 使用接口实现的重试
func Retry(r Retryable, maxAttempts int) error {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := r.Attempt()
		if err == nil {
			return nil
		}

		if !r.ShouldRetry(err) || attempt == maxAttempts {
			return err
		}

		backoff := r.Backoff(attempt)
		fmt.Printf("尝试 %d 失败，%v 后重试...\n", attempt, backoff)
		time.Sleep(backoff)
	}
	return nil
}

type MyOperation struct {
	counter int
}

func (o *MyOperation) Attempt() error {
	o.counter++
	fmt.Printf("执行操作 (第%d次)...\n", o.counter)
	if o.counter < 4 {
		return errors.New("模拟错误")
	}
	return nil
}

func (o *MyOperation) ShouldRetry(err error) bool {
	return true // 总是重试，实际应用中可以根据错误类型判断
}

func (o *MyOperation) Backoff(attempt int) time.Duration {
	return time.Duration(attempt) * time.Second
}

func main() {
	op := &MyOperation{}
	err := Retry(op, 5)
	if err != nil {
		fmt.Println("最终失败:", err)
	} else {
		fmt.Println("操作成功!")
	}
}

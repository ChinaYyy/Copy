package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func UseContext(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Printf("context is done with error %s", ctx.Err())
			return
		default:
			log.Printf("nothing just loop...")
			time.Sleep(time.Second * time.Duration(1))
		}
	}
}

func UseContextFunc() {
	ctx, cancel := context.WithCancel(context.Background())
	go UseContext(ctx)

	time.Sleep(time.Second * time.Duration(1))
	cancel()
	time.Sleep(time.Second * time.Duration(2))
}

func handle(ctx context.Context, duration time.Duration) {
	select {
	case <-ctx.Done():
		fmt.Println("handle", ctx.Err())
	case <-time.After(duration):
		fmt.Println("process request with", duration)
	}
}

func contextTimeoutFunc() {

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go handle(ctx, 500*time.Millisecond)
	select {
	case <-ctx.Done():
		fmt.Println("main", ctx.Err())
	}

}

func main() {
	contextTimeoutFunc()
}

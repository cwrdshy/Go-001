package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"sync"
	"time"
)

func main() {
	//思路：
	//1.main函数作为主goroutine需要等待所有
	//http server服务执行完成返回(waitGroup)
	//2.如果某一服务报错（如两个服务同用1000端口），
	//则快速返回错误，服务1和2都结束执行(errgroup)
	//3.关闭接收错误的channel

	var wg sync.WaitGroup
	wg.Add(2) // 假设有2个http server服务

	ctx, cancel := context.WithCancel(context.Background())

	eg, egCtx := errgroup.WithContext(context.Background())
	eg.Go(firstServer(ctx, &wg))
	eg.Go(secondServer(ctx, &wg))

	go func() {
		<-egCtx.Done()
		cancel()
	}()

	if err := eg.Wait(); err != nil {
		fmt.Printf("Game Over!!!")
		os.Exit(1)
	}
}

func firstServer(ctx context.Context, wg *sync.WaitGroup) func() error {
	return func() error {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`I'm is NO 1 Server`))
		})
		server := &http.Server{Addr: ":1000", Handler: mux}
		errChan := make(chan error, 1)

		go func() {
			<-ctx.Done()
			shutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			if err := server.Shutdown(shutCtx); err != nil {
				errChan <- fmt.Errorf("error from shutdown NO 1 Server: %w\n", err)
			}
			fmt.Println("NO 1 Server closed")
			close(errChan)
			wg.Done()
		}()

		fmt.Println("NO 1 Server starting")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("error from starting NO 1 Server: %w\n", err)
		}
		fmt.Println("NO 1 xxx")
		wg.Wait()

		return <-errChan
	}
}

func secondServer(ctx context.Context, wg *sync.WaitGroup) func() error {
	return func() error {
		mux := http.NewServeMux()

		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`I'm is NO 2 Server`))
		})
		server := &http.Server{Addr: ":1000", Handler: mux}
		errChan := make(chan error, 1)

		go func() {
			<-ctx.Done()
			shutCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			if err := server.Shutdown(shutCtx); err != nil {
				errChan <- fmt.Errorf("error from shutdown NO 2 Server: %w\n", err)
			}
			fmt.Println("NO 2 Server closed")
			close(errChan)
			wg.Done()
		}()

		fmt.Println("NO 2 Server starting")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("error from starting NO 2 Server: %w\n", err)
		}
		fmt.Println("NO 2 xxx")
		wg.Wait()

		return <-errChan
	}
}

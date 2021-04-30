package main

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	group, ctx := errgroup.WithContext(context.Background())

	// linux signal
	group.Go(func() error {
		sig := make(chan os.Signal)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		fmt.Println("signal...")
		select {
		case <- ctx.Done():
			fmt.Println("signal be stopped by http server shutdown")
			close(sig)
			return ctx.Err()
		case cmd := <-sig:
			time.Sleep(time.Second * 3)
			return fmt.Errorf("user exec %v to stop", cmd)
		}
		return nil
	})

	// http server
	group.Go(func() error {
		s := http.Server{
			Addr: ":8080",
			Handler: nil,
		}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			// do something
			fmt.Print("call api\n")
		})

		http.HandleFunc("/close", func(w http.ResponseWriter, r *http.Request) {
			// do something
			s.Shutdown(context.Background())
		})

		fmt.Println("http server ....")

		go func(ctx context.Context) {
			<- ctx.Done()
			fmt.Println("http server stop by linux signal")
			// graceful shutdown server
			s.Shutdown(context.Background())
		}(ctx)

		return s.ListenAndServe()
	})

	if err := group.Wait(); err != nil {
		fmt.Println("error group err: ", err.Error())
	}
	fmt.Println("stop all")
}

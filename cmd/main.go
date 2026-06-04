package main

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	_ "go.uber.org/automaxprocs"
)

func main() {
	defer sentry.Flush(time.Second * 2)

	defer func() {
		// manually capture panic so we can do our own logging
		r := recover()
		if r != nil {
			fmt.Println("------------------", r, string(debug.Stack()))

			defer sentry.Recover()

			panic(r)
		}
	}()

	if err := newApp().Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %+v\n", err)
		os.Exit(1)
	}
}

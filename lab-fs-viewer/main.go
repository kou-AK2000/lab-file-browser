package main

import (
	"fmt"
	"os"

	"linux-fs-viewer/server"
)

func main() {

	if os.Geteuid() == 0 {
		fmt.Println("Do not run as root")
		return
	}

	srv := server.NewServer()

	fmt.Println("Running at http://127.0.0.1:8080")
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

package main

import (
	"fmt"
	"net/http"
	"os"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintf(w, "Hello World")
	if err != nil {
		fmt.Println("runtime error, info: ", err)
	}
	fmt.Println("The current request server port number is: ", os.Args[1])
}

func main() {
	args := os.Args
	addr := fmt.Sprintf(":%v", args[1])
	http.HandleFunc("/hello", HelloHandler)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe error, info: ", err)
	}
}

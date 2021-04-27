package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
)

var count = 1

func PongHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("This is the " + strconv.Itoa(count) + "th visit")
	count++
	_, err := fmt.Fprintf(w, "PONG")
	if err != nil {
		fmt.Println("Runtime error, info: ", err)
	}
	fmt.Println("The current request server port number is: ", os.Args[1])
}

func main() {
	args := os.Args
	addr := fmt.Sprintf(":%v", args[1])
	http.HandleFunc("/ping", PongHandler)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		fmt.Println("ListenAndServe error, info: ", err)
	}
}

package main

import (
	"HulanRiver/src/gateway"
	"flag"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 5000, "Port to serve")
	flag.Parse()

	gateway.Run(port)

}

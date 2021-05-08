package main

import (
	"HulanRiver/src/gateway"
	"flag"
)

var port = flag.Int("port", 5000, "Port to Server")

func main() {
	flag.Parse()

	gateway.Run(*port)

}

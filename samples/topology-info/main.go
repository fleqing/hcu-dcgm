package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/HYGON-AI/hcu-dcgm/v3/pkg/dcgm"
)

func main() {
	flag.Parse()

	if err := dcgm.Init(); err != nil {
		panic(err)
	}
	defer dcgm.ShutDown()

	matrix, err := dcgm.DiscoverInterconnectTopology()
	if err != nil {
		panic(err)
	}
	out, err := json.Marshal(matrix)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}

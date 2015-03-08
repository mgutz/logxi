package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func sendExternal(obj map[string]interface{}) {
	// do something with the log entry here
	fmt.Printf("Sending ... %#v\n", obj)
}

func main() {
	r := bufio.NewReader(os.Stdin)
	dec := json.NewDecoder(r)
	for {
		var obj map[string]interface{}
		if err := dec.Decode(&obj); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		sendExternal(obj)
	}
}

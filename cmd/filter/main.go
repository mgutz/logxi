package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mgutz/logxi"
)

func sendExternal(obj map[string]interface{}) {
	// normally you would send this to an external service like InfluxDB
	// or some logging framework. Let's filter out some data.
	fmt.Printf("Time: %s Level: %s Message: %s\n",
		obj[logxi.KeyMap.Time],
		obj[logxi.KeyMap.Level],
		obj[logxi.KeyMap.Message],
	)
}

func main() {
	r := bufio.NewReader(os.Stdin)
	dec := json.NewDecoder(r)
	for {
		var obj map[string]interface{}
		if err := dec.Decode(&obj); err == io.EOF {
			break
		} else if err != nil {
			logxi.InternalLog.Fatal("Could not decode", "err", err)
		}
		sendExternal(obj)
	}
}

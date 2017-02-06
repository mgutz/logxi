package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type jsonObject map[string]interface{}

// Scans STDIN for newlines. When found, try to JSON Unmarshal the whole line. If not
// valid convert it to the form {"stdin": LINE}
func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Bytes()

		var m jsonObject

		err := json.Unmarshal(line, &m)
		if err == nil {
			fmt.Println(string(line))
		} else {
			b, _ := json.Marshal(jsonObject{"stdin": string(line)})
			fmt.Println(string(b))
		}
	}
}

package main

type jsonObject map[string]interface{}

var _broadcaster = NewBroadcaster(1)

// Scans STDIN for newlines. When found, try to JSON Unmarshal the whole line. If not
// valid convert it to the form {"stdin": LINE}
// func pumpStdout2() {
// 	scanner := bufio.NewScanner(os.Stdin)
// 	for scanner.Scan() {
// 		line := scanner.Bytes()

// 		var m jsonObject

// 		err := json.Unmarshal(line, &m)
// 		if err == nil {
// 			_broadcaster.Send(string(line))
// 		} else {
// 			b, _ := json.Marshal(jsonObject{"stdin": string(line)})
// 			_broadcaster.Send(string(b))
// 			fmt.Println(string(b))
// 		}
// 	}
// }

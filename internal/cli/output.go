package cli

import (
"encoding/json"
"fmt"
)

func PrintResult(result interface{}, config *Config) {
if config.JSON {
data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		switch v := result.(type) {
		case string:
			fmt.Println(v)
		case []map[string]interface{}:
			for _, item := range v {
				printMap(item)
				fmt.Println()
			}
		case map[string]interface{}:
			printMap(v)
		default:
			fmt.Printf("%v\n", v)
		}
	}
}

func printMap(m map[string]interface{}) {
	for k, v := range m {
		fmt.Printf("%s: %v\n", k, v)
	}
}

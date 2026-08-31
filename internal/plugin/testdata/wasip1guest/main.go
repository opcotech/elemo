package main

import (
	"fmt"
	"os"
	"time"

	plugin "github.com/opcotech/elemo/sdk/plugin"
)

func init() {
	plugin.Register(func(req plugin.Request) ([]byte, error) {
		switch req.Function {
		case "start", "stop":
			return plugin.Reply(nil)
		case "ping":
			return plugin.Reply(map[string]any{"pong": true})
		case "now":
			return plugin.Reply(map[string]any{"unix": time.Now().Unix()})
		case "hostping":
			raw, err := plugin.Host("plugin.storage.get", req.ScopeID, map[string]any{"key": "k"})
			if err != nil {
				return nil, err
			}
			return plugin.Reply(map[string]any{"data": string(raw)})
		case "crash":
			os.Exit(4)
			return nil, fmt.Errorf("unreachable")
		default:
			return nil, fmt.Errorf("unknown function %s", req.Function)
		}
	})
}

func main() {}

package signals

import (
	"embed"
	"encoding/json"
)

//go:embed data/agent.json
var agentFS embed.FS

var AgentSignals json.RawMessage

func init() {
	data, err := agentFS.ReadFile("data/agent.json")
	if err != nil {
		panic("failed to load embedded agent.json: " + err.Error())
	}
	AgentSignals = data
}

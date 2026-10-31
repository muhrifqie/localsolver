package signals

import "encoding/json"

type SignalsPayload struct {
	Agent   json.RawMessage `json:"agent"`
	Widget  WidgetSignals   `json:"widget"`
	Trigger TriggerSignals  `json:"trigger"`
	Root    RootSignals     `json:"root"`
	Nonce   uint32          `json:"nonce"`
}

func Generate(wid, sid string, nonce uint32) (string, error) {
	widget := GenerateWidget(wid)
	trigger := GenerateTrigger(widget.T.PNow)
	root := GenerateRoot(wid, sid)

	payload := SignalsPayload{
		Agent:   AgentSignals,
		Widget:  widget,
		Trigger: trigger,
		Root:    root,
		Nonce:   nonce,
	}

	return Encode(payload)
}

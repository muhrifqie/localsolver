package signals

type TriggerSignals struct {
	V     int            `json:"v"`
	Conv  int            `json:"conv"`
	TT    string         `json:"tt"`
	APNow int64          `json:"a_pnow"`
	TPNow int64          `json:"t_pnow"`
	SM    string         `json:"sm"`
	ST    StackTrace     `json:"st"`
	EL    TriggerElement `json:"el"`
	WE    bool           `json:"we"`
	WEIT  bool           `json:"weit"`
	EV    TriggerEvent   `json:"ev"`
}

type TriggerElement struct {
	BCR [4]int `json:"bcr"`
	Con bool   `json:"con"`
}

type TriggerEvent struct {
	TS  float64 `json:"ts"`
	EOT bool    `json:"eot"`
	IT  bool    `json:"it"`
	RT  bool    `json:"rt"`
}

func GenerateTrigger(widgetPNow int64) TriggerSignals {
	return TriggerSignals{
		V:     1,
		Conv:  1,
		TT:    "widget",
		APNow: widgetPNow + 2,
		TPNow: widgetPNow,
		SM:    "",
		ST:    StackTrace{L: 498, N: 3, H: 3268819621},
		EL: TriggerElement{
			BCR: [4]int{0, 0, 316, 70},
			Con: true,
		},
		WE:   true,
		WEIT: true,
		EV: TriggerEvent{
			TS:  float64(widgetPNow) - 0.2,
			EOT: false,
			IT:  true,
			RT:  false,
		},
	}
}

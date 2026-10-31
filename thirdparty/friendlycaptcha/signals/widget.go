package signals

import (
	"math/rand/v2"
	"time"
)

type WidgetSignals struct {
	V    int              `json:"v"`
	Conv int              `json:"conv"`
	I    int              `json:"i"`
	WID  string           `json:"wid"`
	T    WidgetTime       `json:"t"`
	Dims WidgetDims       `json:"dims"`
	Mel  WidgetMouseEnter `json:"mel"`
	ST   StackTrace       `json:"st"`
}

type WidgetTime struct {
	TS   int64 `json:"ts"`
	PTO  int64 `json:"pto"`
	PNow int64 `json:"pnow"`
}

type WidgetDims struct {
	D   [10]int `json:"d"`
	DPR int     `json:"dpr"`
}

type WidgetMouseEnter struct {
	N  int `json:"n"`
	TS int `json:"ts"`
	D  int `json:"d"`
}

type StackTrace struct {
	L int    `json:"l"`
	N int    `json:"n"`
	H uint32 `json:"h"`
}

func GenerateWidget(wid string) WidgetSignals {
	now := time.Now().UnixMilli()
	pnow := int64(800 + rand.IntN(1700))

	return WidgetSignals{
		V:    1,
		Conv: 1,
		I:    2,
		WID:  wid,
		T: WidgetTime{
			TS:   now,
			PTO:  now - pnow,
			PNow: pnow,
		},
		Dims: WidgetDims{
			D:   [10]int{316, 70, 1920, 1032, -1920, 65, 0, 0, 316, 70},
			DPR: 1,
		},
		Mel: WidgetMouseEnter{N: 0, TS: 0, D: 0},
		ST:  StackTrace{L: 746, N: 4, H: 2193565333},
	}
}

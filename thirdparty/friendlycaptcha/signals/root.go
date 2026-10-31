package signals

import (
	"math/rand/v2"
	"time"
)

type RootSignals struct {
	V    int        `json:"v"`
	I    int        `json:"i"`
	HL   int        `json:"hl"`
	FE   bool       `json:"fe"`
	Dep  bool       `json:"dep"`
	WID  string     `json:"wid"`
	SC   int        `json:"sc"`
	SID  string     `json:"sid"`
	Conv int        `json:"conv"`
	T    WidgetTime `json:"t"`
	Dims WidgetDims `json:"dims"`
	Mel  RootMel    `json:"mel"`
	MM   RootMM     `json:"mm"`
	BH   RootBH     `json:"bh"`
	ST   StackTrace `json:"st"`
	PTRC RootPTRC   `json:"ptrc"`
}

type RootMel struct {
	N   int     `json:"n"`
	TS  int     `json:"ts"`
	D   int     `json:"d"`
	FTS int     `json:"fts,omitempty"`
	FXY *[4]int `json:"fxy,omitempty"`
	XY  *[2]int `json:"xy,omitempty"`
}

type RootMM struct {
	XY [10]int `json:"xy"`
	TS int     `json:"ts"`
}

type RootBH struct {
	OnOff map[string][7]int `json:"onoff"`
	NEV   [14]int           `json:"nev"`
	NK    [8]int            `json:"nk"`
	MOV   RootMov           `json:"mov"`
	DM    RootDM            `json:"dm"`
	DO    RootDO            `json:"do"`
}

type RootMov struct {
	T  [7]float64 `json:"t"`
	V  [7]float64 `json:"v"`
	D  [7]float64 `json:"d"`
	NS int        `json:"ns"`
}

type RootDM struct {
	N  int    `json:"n"`
	TS int    `json:"ts"`
	AC [7]int `json:"ac"`
	RR [7]int `json:"rr"`
	I  int    `json:"i"`
	G  bool   `json:"g"`
}

type RootDO struct {
	FTS int    `json:"fts"`
	TS  int    `json:"ts"`
	GD  [7]int `json:"gd"`
	BD  [7]int `json:"bd"`
	A   int    `json:"a"`
	B   int    `json:"b"`
	G   int    `json:"g"`
}

type RootPTRC struct {
	PNow []any `json:"pnow"`
	Src  []any `json:"src"`
	SD   []any `json:"sd"`
	BF   []any `json:"bf"`
}

func GenerateRoot(wid, sid string) RootSignals {
	now := time.Now().UnixMilli()
	pnow := int64(1000 + rand.IntN(2000))

	return RootSignals{
		V:    1,
		I:    2,
		HL:   2,
		FE:   false,
		Dep:  false,
		WID:  wid,
		SC:   2,
		SID:  sid,
		Conv: 1,
		T: WidgetTime{
			TS:   now,
			PTO:  now - pnow,
			PNow: pnow,
		},
		Dims: WidgetDims{
			D:   [10]int{1920, 945, 1920, 1032, -1920, 65, 0, 0, 1904, 70},
			DPR: 1,
		},
		Mel: RootMel{N: 0, TS: 0, D: 0},
		MM:  RootMM{TS: 0},
		BH: RootBH{
			OnOff: map[string][7]int{
				"kdu":  {},
				"cse":  {},
				"mdu":  {},
				"mle":  {1, 0, 0, 0, 0, 0, 0},
				"med":  {},
				"semd": {},
				"se":   {},
				"pdc":  {},
				"mmc":  {},
				"tse":  {},
				"fikd": {},
			},
			NEV: [14]int{5},
			NK:  [8]int{},
			MOV: RootMov{
				T:  [7]float64{1, 96.9, 96.9, 96.9, 0, 0, 0},
				V:  [7]float64{2, 5432.18, 5199.38, 5664.98, 232.80, 0, -2},
				D:  [7]float64{1, 526.50, 526.50, 526.50, 0, 0, 0},
				NS: 1,
			},
			DM: RootDM{},
			DO: RootDO{A: -128},
		},
		ST:   StackTrace{L: 788, N: 9, H: 2189576926},
		PTRC: RootPTRC{PNow: []any{}, Src: []any{}, SD: []any{}, BF: []any{}},
	}
}

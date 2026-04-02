package tdd

import "strings"

type TDDEncoder struct {
	symbols map[string]string
}

func NewTDDEncoder() *TDDEncoder {
	return &TDDEncoder{
		symbols: map[string]string{
			"function":   "fn",
			"variable":   "var",
			"constant":   "const",
			"parameter":  "param",
			"argument":   "arg",
			"return":     "ret",
			"import":     "imp",
			"package":    "pkg",
			"interface":  "iface",
			"struct":     "st",
			"method":     "meth",
			"context":    "ctx",
			"request":    "req",
			"response":   "resp",
			"error":      "err",
			"config":     "cfg",
			"database":   "db",
			"connection": "conn",
			"handler":    "hdlr",
			"middleware": "mw",
		},
	}
}

func (e *TDDEncoder) Encode(input string) string {
	output := input
	for full, short := range e.symbols {
		output = strings.ReplaceAll(output, " "+full+" ", " "+short+" ")
	}
	return output
}

func (e *TDDEncoder) Decode(input string) string {
	output := input
	reversed := make(map[string]string)
	for short, full := range e.symbols {
		reversed[short] = full
	}
	for short, full := range reversed {
		output = strings.ReplaceAll(output, " "+short+" ", " "+full+" ")
	}
	return output
}

func (e *TDDEncoder) CompressionRatio(original, encoded string) float64 {
	if len(original) == 0 {
		return 0
	}
	return float64(len(encoded)) / float64(len(original)) * 100
}

func (e *TDDEncoder) AddSymbol(full, short string) {
	e.symbols[full] = short
}

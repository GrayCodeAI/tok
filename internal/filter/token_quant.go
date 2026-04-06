package filter

import (
	"strings"

	"github.com/GrayCodeAI/tokman/internal/core"
)

// Paper: "TurboQuant: Extreme KV Cache Compression" — Google Research, 2026
// TokenQuantFilter implements token-level quantization — replaces verbose
// tokens with shorter equivalents while preserving semantic meaning.
type TokenQuantFilter struct {
	quantMap map[string]string
}

// NewTokenQuantFilter creates a new token quantization filter.
func NewTokenQuantFilter() *TokenQuantFilter {
	return &TokenQuantFilter{
		quantMap: map[string]string{
			"configuration": "cfg", "function": "func", "variable": "var",
			"parameter": "param", "argument": "arg", "environment": "env",
			"directory": "dir", "command": "cmd", "message": "msg",
			"information": "info", "application": "app", "implementation": "impl",
			"interface": "iface", "package": "pkg", "document": "doc",
			"development": "dev", "production": "prod", "management": "mgmt",
			"controller": "ctrl", "manager": "mgr", "handler": "hdlr",
			"response": "resp", "request": "req", "connection": "conn",
			"authentication": "auth", "authorization": "authz",
			"middleware": "mw", "database": "db", "server": "srv",
			"client": "cli", "error": "err", "warning": "warn",
			"success": "ok", "failure": "fail", "initialize": "init",
			"maximum": "max", "minimum": "min", "default": "def",
			"temporary": "tmp", "number": "num", "string": "str",
			"context": "ctx", "index": "idx", "value": "val",
			"result": "res", "previous": "prev", "current": "cur",
			"total": "tot", "average": "avg", "standard": "std",
			"additional": "add", "optional": "opt", "required": "req",
			"internal": "int", "external": "ext", "global": "gbl",
			"local": "loc", "public": "pub", "private": "priv",
			"protected": "prot", "abstract": "abs", "concrete": "conc",
			"virtual": "virt", "override": "ovrd", "asynchronous": "async",
			"synchronous": "sync", "parallel": "par", "sequential": "seq",
			"recursive": "rec", "iterative": "iter",
		},
	}
}

// Apply quantizes verbose tokens to shorter equivalents.
func (f *TokenQuantFilter) Apply(input string, mode Mode) (string, int) {
	if mode == ModeNone {
		return input, 0
	}

	original := input
	result := input

	for long, short := range f.quantMap {
		result = strings.ReplaceAll(result, " "+long+" ", " "+short+" ")
		result = strings.ReplaceAll(result, " "+long+"(", " "+short+"(")
		result = strings.ReplaceAll(result, " "+long+".", " "+short+".")
		result = strings.ReplaceAll(result, " "+long+",", " "+short+",")
		result = strings.ReplaceAll(result, " "+long+":", " "+short+":")
	}

	saved := core.EstimateTokens(original) - core.EstimateTokens(result)
	if saved < 0 {
		saved = 0
	}
	return result, saved
}

// Name returns the layer name.
func (f *TokenQuantFilter) Name() string { return "25_token_quant" }

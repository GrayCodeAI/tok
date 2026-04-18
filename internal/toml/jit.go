package toml

import "regexp"

// JITCompiler compiles TOML patterns to optimized functions
type JITCompiler struct {
	compiled map[string]func(string) bool
}

func NewJITCompiler() *JITCompiler {
	return &JITCompiler{compiled: make(map[string]func(string) bool)}
}

func (jit *JITCompiler) Compile(pattern string) func(string) bool {
	if fn, ok := jit.compiled[pattern]; ok {
		return fn
	}
	
	re := regexp.MustCompile(pattern)
	fn := func(s string) bool { return re.MatchString(s) }
	jit.compiled[pattern] = fn
	return fn
}

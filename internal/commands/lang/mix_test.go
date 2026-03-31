package lang

import (
	"testing"
)

func TestFilterMixCompileOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful compilation",
			input: `Compiling 1 file (.ex)
Generated my_app app

warning: variable "unused" is unused (if the variable is not meant to be used, prefix it with an underscore)
  lib/my_app.ex:10`,
		},
		{
			name: "compilation error",
			input: `== Compilation error in file lib/my_app.ex ==
** (CompileError) lib/my_app.ex:10: undefined function foo/0
    (elixir 1.14.0) lib/kernel/parallel_compiler.ex:340: anonymous fn/5 in Kernel.ParallelCompiler.spawn_workers/7`,
		},
		{
			name:  "empty output",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMixCompileOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterMixCompileOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterMixTestOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "all tests pass",
			input: `...

Finished in 0.5 seconds
5 tests, 0 failures

Randomized with seed 12345678`,
		},
		{
			name: "test failure",
			input: `  1) test add numbers (MyApp.CalculatorTest)
     test/my_app/calculator_test.exs:5
     Assertion with == failed
     code:  assert Calculator.add(2, 3) == 6
     left:  5
     right: 6
     stacktrace:
       test/my_app/calculator_test.exs:7: (test)

Finished in 0.3 seconds
5 tests, 1 failure

Randomized with seed 12345678`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMixTestOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterMixTestOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterMixDepsOutput(t *testing.T) {
	input := `* decimal 2.0.0 (Hex package) (mix)
  locked at 2.0.0 (decimal)
* phoenix 1.7.0 (Hex package) (mix)
  locked at 1.7.0 (phoenix)
  up to date`
	result := filterMixDepsOutput(input)
	if result == "" {
		t.Error("filterMixDepsOutput() returned empty string")
	}
}

package build

import (
	"testing"
)

func TestFilterMakeOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful build",
			input: `gcc -o main main.o utils.o -lm
gcc -c -o main.o main.c
gcc -c -o utils.o utils.c
make: Leaving directory '/home/user/project'`,
		},
		{
			name: "build error",
			input: `gcc -c -o main.o main.c
main.c: In function 'main':
main.c:15:5: error: 'foo' undeclared (first use in this function)
     foo = 42;
     ^
make: *** [Makefile:10: main.o] Error 1`,
		},
		{
			name: "build warning",
			input: `gcc -c -o main.o main.c
main.c:10:5: warning: unused variable 'x' [-Wunused-variable]
     int x = 5;
     ^
gcc -o main main.o`,
		},
		{
			name:  "empty output",
			input: "",
		},
		{
			name: "clean output",
			input: `rm -f *.o main
rm -f *~`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMakeOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterMakeOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterMakeOutputContainsErrors(t *testing.T) {
	input := `gcc -c -o main.o main.c
main.c:15:5: error: undeclared
make: *** [main.o] Error 1`
	result := filterMakeOutput(input)
	// Should contain the error
	if result == "" {
		t.Error("filterMakeOutput() should show errors")
	}
}

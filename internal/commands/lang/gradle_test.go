package lang

import (
	"testing"
)

func TestFilterGradleBuildOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful build",
			input: `> Task :compileJava
> Task :processResources
> Task :classes
> Task :jar

BUILD SUCCESSFUL in 5s
4 actionable tasks: 4 executed`,
		},
		{
			name: "failed build",
			input: `> Task :compileJava FAILED

error: cannot find symbol
  symbol: variable foo

FAILURE: Build failed with an exception.

* What went wrong:
Execution failed for task ':compileJava'.`,
		},
		{
			name:  "empty output",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGradleBuildOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterGradleBuildOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterGradleTestOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "all tests pass",
			input: `> Task :test

com.example.MyTest > testAdd PASSED
com.example.MyTest > testSubtract PASSED

BUILD SUCCESSFUL in 3s
2 tests completed, 0 failed`,
		},
		{
			name: "test failure",
			input: `> Task :test FAILED

com.example.MyTest > testAdd PASSED
com.example.MyTest > testDivide FAILED

    java.lang.AssertionError: expected:<5> but was:<4>

2 tests completed, 1 failed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterGradleTestOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterGradleTestOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterGradleDependenciesOutput(t *testing.T) {
	input := `compileClasspath - Compile classpath for source set 'main'.
+--- org.springframework.boot:spring-boot-starter-web:2.7.0
|    +--- org.springframework.boot:spring-boot-starter:2.7.0
|    |    +--- org.springframework.boot:spring-boot:2.7.0
|    |    +--- org.springframework.boot:spring-boot-autoconfigure:2.7.0
+--- com.google.guava:guava:31.1-jre`
	result := filterGradleDependenciesOutput(input)
	if result == "" {
		t.Error("filterGradleDependenciesOutput() returned empty string")
	}
}

func TestFilterGradleTasksOutput(t *testing.T) {
	input := `Build tasks
-----------
assemble - Assembles the outputs of this project.
build - Assembles and tests this project.

Verification tasks
------------------
check - Runs all checks.
test - Runs the unit tests.`
	result := filterGradleTasksOutput(input)
	if result == "" {
		t.Error("filterGradleTasksOutput() returned empty string")
	}
}

func TestFilterGradleBasicOutput(t *testing.T) {
	input := "> Task :compileJava\nBUILD SUCCESSFUL in 2s"
	result := filterGradleBasicOutput(input)
	if result == "" {
		t.Error("filterGradleBasicOutput() returned empty string")
	}
}

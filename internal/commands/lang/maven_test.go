package lang

import (
	"testing"
)

func TestFilterMavenBuildOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful build",
			input: `[INFO] --- maven-compiler-plugin:3.8.1:compile (default-compile) @ my-app ---
[INFO] Changes detected - recompiling the module!
[INFO] Compiling 10 source files to /target/classes
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time:  5.123 s`,
		},
		{
			name: "build failure",
			input: `[ERROR] /src/main/java/com/example/App.java:[10,8] error: cannot find symbol
[INFO] ------------------------------------------------------------------------
[INFO] BUILD FAILURE
[INFO] ------------------------------------------------------------------------
[INFO] Total time:  3.456 s`,
		},
		{
			name: "build with warnings",
			input: `[WARNING] /src/main/java/com/example/App.java:[5,12] warning: [deprecation] something
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------`,
		},
		{
			name:  "empty output",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMavenBuildOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterMavenBuildOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterMavenTestOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "all tests pass",
			input: `[INFO] --- maven-surefire-plugin:2.22.2:test (default-test) @ my-app ---
[INFO] -------------------------------------------------------
[INFO]  T E S T S
[INFO] -------------------------------------------------------
[INFO] Running com.example.MyTest
[INFO] Tests run: 5, Failures: 0, Errors: 0, Skipped: 0
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS`,
		},
		{
			name: "test failure",
			input: `[INFO] Running com.example.MyTest
[ERROR] Tests run: 5, Failures: 1, Errors: 0, Skipped: 0 <<< FAILURES!
[ERROR] com.example.MyTest.testAdd <<< FAILURE!
[INFO] ------------------------------------------------------------------------
[INFO] BUILD FAILURE`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterMavenTestOutput(tt.input)
			if result == "" && tt.input != "" {
				t.Error("filterMavenTestOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterMavenDependencyOutput(t *testing.T) {
	input := `[INFO] com.example:my-app:jar:1.0.0
[INFO] +- org.springframework.boot:spring-boot-starter-web:jar:2.7.0:compile
[INFO] |  +- org.springframework.boot:spring-boot-starter:jar:2.7.0:compile
[INFO] |  |  \- org.springframework.boot:spring-boot:jar:2.7.0:compile
[INFO] \- com.google.guava:guava:jar:31.1-jre:compile`
	result := filterMavenDependencyOutput(input)
	if result == "" {
		t.Error("filterMavenDependencyOutput() returned empty string")
	}
}

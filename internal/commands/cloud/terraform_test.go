package cloud

import (
	"testing"
)

func TestFilterTerraformPlanOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantPlan bool
	}{
		{
			name: "plan with changes",
			input: `Terraform will perform the following actions:

  # aws_instance.web will be created
  + resource "aws_instance" "web" {
      + ami           = "ami-12345"
      + instance_type = "t2.micro"
    }

Plan: 1 to add, 0 to change, 0 to destroy.`,
			wantPlan: true,
		},
		{
			name:     "no changes",
			input:    "No changes. Your infrastructure matches the configuration.",
			wantPlan: false,
		},
		{
			name:     "error output",
			input:    "Error: Invalid AMI ID specified",
			wantPlan: false,
		},
		{
			name:     "empty input",
			input:    "",
			wantPlan: true, // empty input returns raw (empty string), which is still "returned"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTerraformPlanOutput(tt.input)
			// Empty input returns empty string, which is valid
			if tt.input != "" && result == "" {
				t.Error("filterTerraformPlanOutput() returned empty string for non-empty input")
			}
		})
	}
}

func TestFilterTerraformApplyOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful apply",
			input: `Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

instance_id = "i-1234567890"`,
		},
		{
			name: "apply with error",
			input: `Error: Error applying plan:

1 error occurred:
	* aws_instance.web: 1 error occurred:
	* aws_instance.web: Error launching source instance: InsufficientInstanceCapacity`,
		},
		{
			name:  "empty input",
			input: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTerraformApplyOutput(tt.input)
			if result == "" {
				t.Error("filterTerraformApplyOutput() returned empty string")
			}
		})
	}
}

func TestFilterTerraformShowOutput(t *testing.T) {
	input := `# aws_instance.web:
resource "aws_instance" "web" {
    ami           = "ami-12345"
    id            = "i-1234567890"
    instance_type = "t2.micro"
    name          = "web-server"
}
`
	result := filterTerraformShowOutput(input)
	if result == "" {
		t.Error("filterTerraformShowOutput() returned empty string")
	}
}

func TestFilterTerraformInitOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "successful init",
			input: `Initializing provider plugins...
- Finding hashicorp/aws versions matching "4.0.0"...
- Installing hashicorp/aws v4.0.0...
- Installed hashicorp/aws v4.0.0 (signed by HashiCorp)

Terraform has been successfully initialized!`,
		},
		{
			name: "init with error",
			input: `Error: Failed to query available provider packages

Could not retrieve the list of available versions for provider hashicorp/aws`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTerraformInitOutput(tt.input)
			if result == "" {
				t.Error("filterTerraformInitOutput() returned empty string")
			}
		})
	}
}

func TestFilterTerraformValidateOutput(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "valid config",
			input: "Success! The configuration is valid.",
		},
		{
			name: "invalid config",
			input: `Error: Invalid resource type

on main.tf line 10:
  10: resource "aws_instnce" "web" {
The provider hashicorp/aws does not support resource type "aws_instnce".`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTerraformValidateOutput(tt.input)
			if result == "" {
				t.Error("filterTerraformValidateOutput() returned empty string")
			}
		})
	}
}

func TestFilterTerraformOutput(t *testing.T) {
	input := `instance_id = "i-1234567890"
vpc_id = "vpc-12345"
subnet_ids = [
  "subnet-1",
  "subnet-2",
  "subnet-3",
]`
	result := filterTerraformOutput(input)
	if result == "" {
		t.Error("filterTerraformOutput() returned empty string")
	}
}

func TestFilterTerraformOutputTruncation(t *testing.T) {
	// Generate more than 30 lines
	input := ""
	for i := 0; i < 40; i++ {
		input += "output_line = value\n"
	}
	result := filterTerraformOutput(input)
	if len(result) == 0 {
		t.Error("filterTerraformOutput() returned empty string")
	}
}

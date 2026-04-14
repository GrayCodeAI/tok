package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tokman/internal/commands/registry"
	"github.com/GrayCodeAI/tokman/internal/commands/shared"
	"github.com/GrayCodeAI/tokman/internal/filter"
	"github.com/GrayCodeAI/tokman/internal/tracking"
)

var awsCmd = &cobra.Command{
	Use:   "aws [service] [command]",
	Short: "AWS CLI with compact output",
	Long: `AWS CLI commands with compact JSON output and filtering.

Supports compact output for common AWS services:
  sts        - Security Token Service (get-caller-identity)
  s3         - S3 operations (ls, cp, sync)
  ec2        - EC2 operations (describe-instances, describe-vpcs)
  ecs        - ECS operations (list-clusters, describe-services)
  rds        - RDS operations (describe-db-instances)
  lambda     - Lambda operations (list-functions, get-function)
  cloudformation - CloudFormation stacks
  dynamodb   - DynamoDB tables
  iam        - IAM users, roles, policies
  logs       - CloudWatch Logs
  sns        - SNS topics
  sqs        - SQS queues
  kms        - KMS keys
  secretsmanager - Secrets Manager
  ssm        - Systems Manager Parameter Store

Examples:
  tokman aws sts get-caller-identity
  tokman aws s3 ls
  tokman aws ec2 describe-instances
  tokman aws lambda list-functions`,
	DisableFlagParsing: true,
	RunE:               runAws,
}

func init() {
	registry.Add(func() { registry.Register(awsCmd) })
}

func runAws(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}

	timer := tracking.Start()

	if shared.Verbose > 0 {
		fmt.Fprintf(os.Stderr, "Running: aws %s\n", strings.Join(args, " "))
	}

	// Force JSON output
	awsArgs := append([]string{"--output", "json"}, args...)

	execCmd := exec.Command("aws", awsArgs...)
	output, err := execCmd.CombinedOutput()
	raw := string(output)

	// Check if output is JSON
	var filtered string
	if isJSON(raw) {
		filtered = filterAwsJSON(raw, args)
	} else {
		filtered = filterAwsText(raw)
	}

	fmt.Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("aws %s", strings.Join(args, " ")), "tokman aws", originalTokens, filteredTokens)

	return err
}

func isJSON(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")
}

func filterAwsJSON(raw string, args []string) string {
	// Parse JSON
	var data interface{}
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return raw
	}

	// Determine service from args
	service := ""
	if len(args) > 0 {
		service = args[0]
	}

	// Compact based on service
	switch service {
	case "sts":
		return compactSTS(data)
	case "s3":
		return compactS3(data, args)
	case "ec2":
		return compactEC2(data, args)
	case "ecs":
		return compactECS(data, args)
	case "rds":
		return compactRDS(data, args)
	case "lambda":
		return compactLambda(data, args)
	case "cloudformation":
		return compactCloudFormation(data, args)
	case "dynamodb":
		return compactDynamoDB(data, args)
	case "iam":
		return compactIAM(data, args)
	case "logs":
		return compactLogs(data, args)
	case "sns":
		return compactSNS(data, args)
	case "sqs":
		return compactSQS(data, args)
	case "kms":
		return compactKMS(data, args)
	case "secretsmanager":
		return compactSecretsManager(data, args)
	case "ssm":
		return compactSSM(data, args)
	default:
		// Generic JSON compaction
		return compactGeneric(data)
	}
}

func compactSTS(data interface{}) string {
	if m, ok := data.(map[string]interface{}); ok {
		var parts []string
		if arn, ok := m["Arn"].(string); ok {
			parts = append(parts, fmt.Sprintf("ARN: %s", arn))
		}
		if account, ok := m["Account"].(string); ok {
			parts = append(parts, fmt.Sprintf("Account: %s", account))
		}
		if user, ok := m["UserId"].(string); ok {
			parts = append(parts, fmt.Sprintf("UserID: %s", user))
		}
		return strings.Join(parts, "\n")
	}
	return formatGeneric(data)
}

func compactS3(data interface{}, args []string) string {
	// Check if this is ls output
	if arr, ok := data.([]interface{}); ok {
		var items []string
		for _, item := range arr {
			if m, ok := item.(map[string]interface{}); ok {
				if key, ok := m["Key"].(string); ok {
					size := ""
					if s, ok := m["Size"].(float64); ok {
						size = formatBytes(int64(s))
					}
					items = append(items, fmt.Sprintf("  %s (%s)", key, size))
				} else if name, ok := m["Name"].(string); ok {
					items = append(items, fmt.Sprintf("  %s/", name))
				}
			}
		}
		if len(items) > 0 {
			return fmt.Sprintf("📁 S3 Objects (%d):\n%s", len(items), strings.Join(items, "\n"))
		}
	}
	return formatGeneric(data)
}

func compactEC2(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if reservations, ok := m["Reservations"].([]interface{}); ok {
			var instances []string
			for _, r := range reservations {
				if rm, ok := r.(map[string]interface{}); ok {
					if insts, ok := rm["Instances"].([]interface{}); ok {
						for _, i := range insts {
							if im, ok := i.(map[string]interface{}); ok {
								id := ""
								if idVal, ok := im["InstanceId"].(string); ok {
									id = idVal
								}
								state := ""
								if s, ok := im["State"].(map[string]interface{}); ok {
									if name, ok := s["Name"].(string); ok {
										state = name
									}
								}
								itype := ""
								if t, ok := im["InstanceType"].(string); ok {
									itype = t
								}
								instances = append(instances, fmt.Sprintf("  %s (%s, %s)", id, itype, state))
							}
						}
					}
				}
			}
			if len(instances) > 0 {
				return fmt.Sprintf("🖥️  EC2 Instances (%d):\n%s", len(instances), strings.Join(instances, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactECS(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if clusters, ok := m["clusterArns"].([]interface{}); ok {
			var items []string
			for _, c := range clusters {
				if arn, ok := c.(string); ok {
					parts := strings.Split(arn, "/")
					name := arn
					if len(parts) > 1 {
						name = parts[len(parts)-1]
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("🐳 ECS Clusters (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		if services, ok := m["services"].([]interface{}); ok {
			var items []string
			for _, s := range services {
				if sm, ok := s.(map[string]interface{}); ok {
					name := ""
					if n, ok := sm["serviceName"].(string); ok {
						name = n
					}
					status := ""
					if st, ok := sm["status"].(string); ok {
						status = st
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, status))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("🐳 ECS Services (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactRDS(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if instances, ok := m["DBInstances"].([]interface{}); ok {
			var items []string
			for _, i := range instances {
				if im, ok := i.(map[string]interface{}); ok {
					id := ""
					if val, ok := im["DBInstanceIdentifier"].(string); ok {
						id = val
					}
					engine := ""
					if val, ok := im["Engine"].(string); ok {
						engine = val
					}
					status := ""
					if val, ok := im["DBInstanceStatus"].(string); ok {
						status = val
					}
					items = append(items, fmt.Sprintf("  %s (%s, %s)", id, engine, status))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("🗄️  RDS Instances (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactLambda(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if functions, ok := m["Functions"].([]interface{}); ok {
			var items []string
			for _, f := range functions {
				if fm, ok := f.(map[string]interface{}); ok {
					name := ""
					if val, ok := fm["FunctionName"].(string); ok {
						name = val
					}
					runtime := ""
					if val, ok := fm["Runtime"].(string); ok {
						runtime = val
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, runtime))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("λ Lambda Functions (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactCloudFormation(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if stacks, ok := m["Stacks"].([]interface{}); ok {
			var items []string
			for _, s := range stacks {
				if sm, ok := s.(map[string]interface{}); ok {
					name := ""
					if val, ok := sm["StackName"].(string); ok {
						name = val
					}
					status := ""
					if val, ok := sm["StackStatus"].(string); ok {
						status = val
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, status))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📚 CloudFormation Stacks (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactDynamoDB(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if tables, ok := m["TableNames"].([]interface{}); ok {
			var items []string
			for _, t := range tables {
				if name, ok := t.(string); ok {
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📊 DynamoDB Tables (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		if desc, ok := m["Table"].(map[string]interface{}); ok {
			name := ""
			if val, ok := desc["TableName"].(string); ok {
				name = val
			}
			status := ""
			if val, ok := desc["TableStatus"].(string); ok {
				status = val
			}
			count := 0
			if val, ok := desc["ItemCount"].(float64); ok {
				count = int(val)
			}
			return fmt.Sprintf("📊 Table: %s (%s, %d items)", name, status, count)
		}
	}
	return formatGeneric(data)
}

func compactIAM(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		// Users
		if users, ok := m["Users"].([]interface{}); ok {
			var items []string
			for _, u := range users {
				if um, ok := u.(map[string]interface{}); ok {
					name := ""
					if val, ok := um["UserName"].(string); ok {
						name = val
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("👤 IAM Users (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		// Roles
		if roles, ok := m["Roles"].([]interface{}); ok {
			var items []string
			for _, r := range roles {
				if rm, ok := r.(map[string]interface{}); ok {
					name := ""
					if val, ok := rm["RoleName"].(string); ok {
						name = val
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("🔑 IAM Roles (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		// Policies
		if policies, ok := m["Policies"].([]interface{}); ok {
			var items []string
			for _, p := range policies {
				if pm, ok := p.(map[string]interface{}); ok {
					name := ""
					if val, ok := pm["PolicyName"].(string); ok {
						name = val
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📋 IAM Policies (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactLogs(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if groups, ok := m["logGroups"].([]interface{}); ok {
			var items []string
			for _, g := range groups {
				if gm, ok := g.(map[string]interface{}); ok {
					name := ""
					if val, ok := gm["logGroupName"].(string); ok {
						name = val
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📝 CloudWatch Log Groups (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		// Log events
		if events, ok := m["events"].([]interface{}); ok {
			var items []string
			for _, e := range events {
				if em, ok := e.(map[string]interface{}); ok {
					msg := ""
					if val, ok := em["message"].(string); ok {
						msg = val
					}
					if len(msg) > 100 {
						msg = msg[:97] + "..."
					}
					items = append(items, msg)
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📝 Log Events (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactSNS(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if topics, ok := m["Topics"].([]interface{}); ok {
			var items []string
			for _, t := range topics {
				if tm, ok := t.(map[string]interface{}); ok {
					arn := ""
					if val, ok := tm["TopicArn"].(string); ok {
						arn = val
					}
					parts := strings.Split(arn, ":")
					name := arn
					if len(parts) > 0 {
						name = parts[len(parts)-1]
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📢 SNS Topics (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactSQS(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if urls, ok := m["QueueUrls"].([]interface{}); ok {
			var items []string
			for _, u := range urls {
				if url, ok := u.(string); ok {
					parts := strings.Split(url, "/")
					name := url
					if len(parts) > 0 {
						name = parts[len(parts)-1]
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("📨 SQS Queues (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactKMS(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if keys, ok := m["Keys"].([]interface{}); ok {
			return fmt.Sprintf("🔐 KMS Keys: %d", len(keys))
		}
		if desc, ok := m["KeyMetadata"].(map[string]interface{}); ok {
			id := ""
			if val, ok := desc["KeyId"].(string); ok {
				id = val
			}
			descVal := ""
			if val, ok := desc["Description"].(string); ok {
				descVal = val
			}
			return fmt.Sprintf("🔐 Key: %s (%s)", id, descVal)
		}
	}
	return formatGeneric(data)
}

func compactSecretsManager(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if secrets, ok := m["SecretList"].([]interface{}); ok {
			var items []string
			for _, s := range secrets {
				if sm, ok := s.(map[string]interface{}); ok {
					name := ""
					if val, ok := sm["Name"].(string); ok {
						name = val
					}
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("🔒 Secrets (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactSSM(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if params, ok := m["Parameters"].([]interface{}); ok {
			var items []string
			for _, p := range params {
				if pm, ok := p.(map[string]interface{}); ok {
					name := ""
					if val, ok := pm["Name"].(string); ok {
						name = val
					}
					typeVal := ""
					if val, ok := pm["Type"].(string); ok {
						typeVal = val
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, typeVal))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("⚙️  SSM Parameters (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactGeneric(data interface{}) string {
	return formatGeneric(data)
}

func formatGeneric(data interface{}) string {
	// Pretty print JSON with indentation
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(data); err != nil {
		return fmt.Sprintf("%v", data)
	}
	return buf.String()
}

func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func filterAwsText(raw string) string {
	lines := strings.Split(raw, "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}

	if len(result) > 50 {
		return strings.Join(result[:50], "\n") + fmt.Sprintf("\n... (%d more lines)", len(result)-50)
	}
	return strings.Join(result, "\n")
}

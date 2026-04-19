package cloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	out "github.com/lakshmanpatel/tok/internal/output"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lakshmanpatel/tok/internal/commands/registry"
	"github.com/lakshmanpatel/tok/internal/commands/shared"
	"github.com/lakshmanpatel/tok/internal/filter"
	"github.com/lakshmanpatel/tok/internal/tracking"
)

var awsCmd = &cobra.Command{
	Use:   "aws [service] [command]",
	Short: "AWS CLI with compact output",
	Long: `AWS CLI commands with compact JSON output and filtering.

Supports compact output for common AWS services:
  sts        - Security Token Service (get-caller-identity)
  s3         - S3 operations (ls, cp, sync)
  ec2        - EC2 operations (describe-instances, describe-vpcs, describe-security-groups, describe-subnets)
  ecs        - ECS operations (list-clusters, describe-services, list-tasks)
  ecr        - ECR operations (describe-repositories, list-images)
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
  elbv2      - ELBv2 load balancers
  apigateway - API Gateway
  cloudfront - CloudFront distributions
  route53    - Route53 hosted zones and records
  eks        - EKS clusters
  beanstalk  - Elastic Beanstalk environments
  sfn        - Step Functions state machines
  elasticsearch - Elasticsearch domains
  kinesis    - Kinesis streams

Examples:
  tok aws sts get-caller-identity
  tok aws s3 ls
  tok aws ec2 describe-instances
  tok aws lambda list-functions
  tok aws ec2 describe-security-groups
  tok aws route53 list-hosted-zones`,
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
		out.Global().Errorf("Running: aws %s\n", strings.Join(args, " "))
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

	out.Global().Println(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("aws %s", strings.Join(args, " ")), "tok aws", originalTokens, filteredTokens)

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
	case "ecr":
		return compactECR(data, args)
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
	case "elbv2":
		return compactELBv2(data, args)
	case "apigateway":
		return compactAPIGateway(data, args)
	case "cloudfront":
		return compactCloudFront(data, args)
	case "route53":
		return compactRoute53(data, args)
	case "eks":
		return compactEKS(data, args)
	case "elasticbeanstalk":
		return compactBeanstalk(data, args)
	case "stepfunctions", "sfn":
		return compactStepFunctions(data, args)
	case "es", "elasticsearch":
		return compactElasticsearch(data, args)
	case "kinesis":
		return compactKinesis(data, args)
	default:
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
			return fmt.Sprintf("S3 Objects (%d):\n%s", len(items), strings.Join(items, "\n"))
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
								compactEC2Instance(im, &instances)
							}
						}
					}
				}
			}
			if len(instances) > 0 {
				return fmt.Sprintf("EC2 Instances (%d):\n%s", len(instances), strings.Join(instances, "\n"))
			}
		}
		if securityGroups, ok := m["SecurityGroups"].([]interface{}); ok {
			var sgs []string
			for _, sg := range securityGroups {
				if sgm, ok := sg.(map[string]interface{}); ok {
					id, _ := sgm["GroupId"].(string)
					name, _ := sgm["GroupName"].(string)
					desc, _ := sgm["Description"].(string)
					if len(desc) > 50 {
						desc = desc[:47] + "..."
					}
					sgs = append(sgs, fmt.Sprintf("  %s %s - %s", id, name, desc))
				}
			}
			if len(sgs) > 0 {
				return fmt.Sprintf("Security Groups (%d):\n%s", len(sgs), strings.Join(sgs, "\n"))
			}
		}
		if subnets, ok := m["Subnets"].([]interface{}); ok {
			var sns []string
			for _, sn := range subnets {
				if snm, ok := sn.(map[string]interface{}); ok {
					id, _ := snm["SubnetId"].(string)
					az, _ := snm["AvailabilityZone"].(string)
					cidr, _ := snm["CidrBlock"].(string)
					availIPs := ""
					if avm, ok := snm["AvailableIpAddressCount"].(float64); ok {
						availIPs = fmt.Sprintf(" (%d IPs)", int(avm))
					}
					sns = append(sns, fmt.Sprintf("  %s %s %s%s", id, az, cidr, availIPs))
				}
			}
			if len(sns) > 0 {
				return fmt.Sprintf("Subnets (%d):\n%s", len(sns), strings.Join(sns, "\n"))
			}
		}
		if vpcs, ok := m["Vpcs"].([]interface{}); ok {
			var vlist []string
			for _, v := range vpcs {
				if vm, ok := v.(map[string]interface{}); ok {
					id, _ := vm["VpcId"].(string)
					cidr, _ := vm["CidrBlock"].(string)
					state, _ := vm["State"].(string)
					isDefault := ""
					if dv, ok := vm["IsDefault"].(bool); ok && dv {
						isDefault = " [default]"
					}
					vlist = append(vlist, fmt.Sprintf("  %s %s (%s)%s", id, cidr, state, isDefault))
				}
			}
			if len(vlist) > 0 {
				return fmt.Sprintf("VPCs (%d):\n%s", len(vlist), strings.Join(vlist, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactEC2Instance(im map[string]interface{}, instances *[]string) {
	id, _ := im["InstanceId"].(string)
	state := ""
	if s, ok := im["State"].(map[string]interface{}); ok {
		state, _ = s["Name"].(string)
	}
	itype, _ := im["InstanceType"].(string)
	az, _ := im["AvailabilityZone"].(string)
	name := ""
	if tags, ok := im["Tags"].([]interface{}); ok {
		for _, t := range tags {
			if tm, ok := t.(map[string]interface{}); ok {
				if k, _ := tm["Key"].(string); k == "Name" {
					name, _ = tm["Value"].(string)
				}
			}
		}
	}
	if shared.UltraCompact {
		*instances = append(*instances, fmt.Sprintf("%s(%s/%s)", id, itype, state))
	} else {
		label := id
		if name != "" {
			label = fmt.Sprintf("%s/%s", name, id)
		}
		*instances = append(*instances, fmt.Sprintf("  %s (%s, %s, %s)", label, itype, state, az))
	}
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
				return fmt.Sprintf("ECS Clusters (%d):\n%s", len(items), strings.Join(items, "\n"))
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
				return fmt.Sprintf("ECS Services (%d):\n%s", len(items), strings.Join(items, "\n"))
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
				return fmt.Sprintf("RDS Instances (%d):\n%s", len(items), strings.Join(items, "\n"))
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
				return fmt.Sprintf("DynamoDB Tables (%d):\n%s", len(items), strings.Join(items, "\n"))
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
			return fmt.Sprintf("Table: %s (%s, %d items)", name, status, count)
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
				return fmt.Sprintf("IAM Policies (%d):\n%s", len(items), strings.Join(items, "\n"))
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
				return fmt.Sprintf("CloudWatch Log Groups (%d):\n%s", len(items), strings.Join(items, "\n"))
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
				return fmt.Sprintf("Log Events (%d):\n%s", len(items), strings.Join(items, "\n"))
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
			return fmt.Sprintf("KMS Keys: %d", len(keys))
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
			return fmt.Sprintf("Key: %s (%s)", id, descVal)
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
				return fmt.Sprintf("Secrets (%d):\n%s", len(items), strings.Join(items, "\n"))
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
				return fmt.Sprintf("SSM Parameters (%d):\n%s", len(items), strings.Join(items, "\n"))
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

func compactECR(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if repos, ok := m["repositories"].([]interface{}); ok {
			var items []string
			for _, r := range repos {
				if rm, ok := r.(map[string]interface{}); ok {
					name, _ := rm["repositoryName"].(string)
					uri, _ := rm["repositoryUri"].(string)
					if len(uri) > 50 {
						uri = "..." + uri[len(uri)-47:]
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, uri))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("ECR Repositories (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		if imageIds, ok := m["imageIds"].([]interface{}); ok {
			var items []string
			for _, img := range imageIds {
				if im, ok := img.(map[string]interface{}); ok {
					tag, _ := im["imageTag"].(string)
					digest, _ := im["imageDigest"].(string)
					if len(digest) > 19 {
						digest = digest[:19] + "..."
					}
					items = append(items, fmt.Sprintf("  %s (%s)", tag, digest))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("ECR Images (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactELBv2(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if lbs, ok := m["LoadBalancers"].([]interface{}); ok {
			var items []string
			for _, lb := range lbs {
				if lbm, ok := lb.(map[string]interface{}); ok {
					name, _ := lbm["LoadBalancerName"].(string)
					scheme, _ := lbm["Scheme"].(string)
					lbType, _ := lbm["Type"].(string)
					state := ""
					if sm, ok := lbm["State"].(map[string]interface{}); ok {
						state, _ = sm["Code"].(string)
					}
					items = append(items, fmt.Sprintf("  %s [%s/%s] %s", name, scheme, lbType, state))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Load Balancers (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		if tgts, ok := m["TargetGroups"].([]interface{}); ok {
			var items []string
			for _, tg := range tgts {
				if tgm, ok := tg.(map[string]interface{}); ok {
					name, _ := tgm["TargetGroupName"].(string)
					arn, _ := tgm["TargetGroupArn"].(string)
					if len(arn) > 30 {
						arn = "..." + arn[len(arn)-27:]
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, arn))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Target Groups (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactAPIGateway(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if apis, ok := m["items"].([]interface{}); ok {
			var items []string
			for _, api := range apis {
				if am, ok := api.(map[string]interface{}); ok {
					name, _ := am["name"].(string)
					id, _ := am["id"].(string)
					items = append(items, fmt.Sprintf("  %s (%s)", name, id))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("API Gateways (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactCloudFront(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if distList, ok := m["DistributionList"].(map[string]interface{}); ok {
			if items, ok := distList["Items"].([]interface{}); ok {
				var dists []string
				for _, d := range items {
					if dm, ok := d.(map[string]interface{}); ok {
						id, _ := dm["Id"].(string)
						domain, _ := dm["DomainName"].(string)
						comment, _ := dm["Comment"].(string)
						enabled := ""
						if e, ok := dm["Enabled"].(bool); ok {
							if e {
								enabled = "enabled"
							} else {
								enabled = "disabled"
							}
						}
						status, _ := dm["Status"].(string)
						label := id
						if comment != "" && len(comment) < 40 {
							label = comment
						}
						dists = append(dists, fmt.Sprintf("  %s (%s) %s [%s]", label, domain, status, enabled))
					}
				}
				if len(dists) > 0 {
					return fmt.Sprintf("CloudFront Distributions (%d):\n%s", len(dists), strings.Join(dists, "\n"))
				}
			}
		}
	}
	return formatGeneric(data)
}

func compactRoute53(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if hzs, ok := m["HostedZones"].([]interface{}); ok {
			var items []string
			for _, hz := range hzs {
				if hzm, ok := hz.(map[string]interface{}); ok {
					name, _ := hzm["Name"].(string)
					id, _ := hzm["Id"].(string)
					if len(id) > 15 {
						id = "..." + id[len(id)-12:]
					}
					recordCount := ""
					if rc, ok := hzm["ResourceRecordSetCount"].(float64); ok {
						recordCount = fmt.Sprintf(" (%d records)", int(rc))
					}
					items = append(items, fmt.Sprintf("  %s %s%s", name, id, recordCount))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Route53 Hosted Zones (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		if rrsList, ok := m["ResourceRecordSets"].([]interface{}); ok {
			var items []string
			for _, rrs := range rrsList {
				if rrm, ok := rrs.(map[string]interface{}); ok {
					name, _ := rrm["Name"].(string)
					rtype, _ := rrm["Type"].(string)
					ttl, _ := rrm["TTL"].(float64)
					items = append(items, fmt.Sprintf("  %s %s TTL=%d", name, rtype, int(ttl)))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Route53 Records (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactEKS(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if clusters, ok := m["clusters"].([]interface{}); ok {
			var items []string
			for _, c := range clusters {
				if cm, ok := c.(map[string]interface{}); ok {
					name, _ := cm["name"].(string)
					version, _ := cm["version"].(string)
					status, _ := cm["status"].(string)
					items = append(items, fmt.Sprintf("  %s (v%s, %s)", name, version, status))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("EKS Clusters (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactBeanstalk(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if envs, ok := m["Environments"].([]interface{}); ok {
			var items []string
			for _, e := range envs {
				if em, ok := e.(map[string]interface{}); ok {
					name, _ := em["EnvironmentName"].(string)
					status, _ := em["Status"].(string)
					health, _ := em["Health"].(string)
					items = append(items, fmt.Sprintf("  %s [%s, %s]", name, health, status))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Beanstalk Environments (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactStepFunctions(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if stateMachines, ok := m["stateMachines"].([]interface{}); ok {
			var items []string
			for _, sm := range stateMachines {
				if smm, ok := sm.(map[string]interface{}); ok {
					name, _ := smm["name"].(string)
					arn, _ := smm["stateMachineArn"].(string)
					if len(arn) > 30 {
						arn = "..." + arn[len(arn)-27:]
					}
					items = append(items, fmt.Sprintf("  %s (%s)", name, arn))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Step Functions (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactElasticsearch(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if domains, ok := m["DomainNames"].([]interface{}); ok {
			var items []string
			for _, d := range domains {
				if dm, ok := d.(map[string]interface{}); ok {
					name, _ := dm["DomainName"].(string)
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Elasticsearch Domains (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
	}
	return formatGeneric(data)
}

func compactKinesis(data interface{}, args []string) string {
	if m, ok := data.(map[string]interface{}); ok {
		if names, ok := m["StreamNames"].([]interface{}); ok {
			var items []string
			for _, n := range names {
				if name, ok := n.(string); ok {
					items = append(items, fmt.Sprintf("  %s", name))
				}
			}
			if len(items) > 0 {
				return fmt.Sprintf("Kinesis Streams (%d):\n%s", len(items), strings.Join(items, "\n"))
			}
		}
		if desc, ok := m["StreamDescriptionSummary"].(map[string]interface{}); ok {
			name, _ := desc["StreamName"].(string)
			status, _ := desc["StreamStatus"].(string)
			shards := 0
			if s, ok := desc["OpenShardCount"].(float64); ok {
				shards = int(s)
			}
			return fmt.Sprintf("Kinesis Stream: %s (%s, %d shards)", name, status, shards)
		}
	}
	return formatGeneric(data)
}

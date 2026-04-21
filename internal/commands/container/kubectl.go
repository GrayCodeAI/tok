package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	out "github.com/GrayCodeAI/tok/internal/output"

	"github.com/spf13/cobra"

	"github.com/GrayCodeAI/tok/internal/commands/registry"
	"github.com/GrayCodeAI/tok/internal/commands/shared"
	"github.com/GrayCodeAI/tok/internal/filter"
	"github.com/GrayCodeAI/tok/internal/tracking"
)

var kubectlJSON bool

func formatAsJSONKubectl(output string) string {
	return fmt.Sprintf(`{"output": %s}`, strconv.Quote(output))
}

var kubectlCmd = &cobra.Command{
	Use:   "kubectl [command] [args...]",
	Short: "Kubernetes CLI with filtered output",
	Long: `Kubernetes CLI with token-optimized output.

Specialized filters for common commands:
  - kubectl get pods: Pod status summary
  - kubectl get services: Service listing
  - kubectl get deployments: Deployment status
  - kubectl get nodes: Node status
  - kubectl get namespaces: Namespace listing
  - kubectl get configmaps: ConfigMap listing
  - kubectl get secrets: Secret listing
  - kubectl get ingress: Ingress rules
  - kubectl get events: Event summary
  - kubectl logs: Log deduplication
  - kubectl describe: Compact describe output
  - kubectl top pods/nodes: Resource usage

Examples:
  tok kubectl get pods
  tok kubectl get pods -n production
  tok kubectl get services
  tok kubectl logs <pod>
  tok kubectl describe pod <name>
  tok kubectl top pods`,
	DisableFlagParsing: true,
	RunE:               runKubectl,
}

func init() {
	registry.Add(func() { registry.Register(kubectlCmd) })
	kubectlCmd.Flags().BoolVarP(&kubectlJSON, "json", "j", false, "Output as JSON")
}

func runKubectl(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return runKubectlPassthrough(args)
	}

	if args[0] == "get" && len(args) > 1 {
		resourceType := args[1]
		remaining := args[2:]
		switch resourceType {
		case "pods", "pod", "po":
			return runKubectlPods(remaining)
		case "services", "service", "svc":
			return runKubectlServices(remaining)
		case "deployments", "deployment", "deploy":
			return runKubectlDeployments(remaining)
		case "nodes", "node", "no":
			return runKubectlNodes(remaining)
		case "namespaces", "namespace", "ns":
			return runKubectlNamespaces(remaining)
		case "configmaps", "configmap", "cm":
			return runKubectlConfigMaps(remaining)
		case "secrets", "secret":
			return runKubectlSecrets(remaining)
		case "ingress", "ing", "ingresses":
			return runKubectlIngress(remaining)
		case "events", "event", "ev":
			return runKubectlEvents(remaining)
		case "jobs", "job":
			return runKubectlJobs(remaining)
		case "cronjobs", "cronjob", "cj":
			return runKubectlCronJobs(remaining)
		case "persistentvolumes", "persistentvolume", "pv":
			return runKubectlPVs(remaining)
		case "persistentvolumeclaims", "persistentvolumeclaim", "pvc":
			return runKubectlPVCs(remaining)
		case "serviceaccounts", "serviceaccount", "sa":
			return runKubectlServiceAccounts(remaining)
		case "replicasets", "replicaset", "rs":
			return runKubectlReplicaSets(remaining)
		case "daemonsets", "daemonset", "ds":
			return runKubectlDaemonSets(remaining)
		case "statefulsets", "statefulset", "sts":
			return runKubectlStatefulSets(remaining)
		case "horizontalpodautoscalers", "horizontalpodautoscaler", "hpa":
			return runKubectlHPA(remaining)
		case "networkpolicies", "networkpolicy", "netpol":
			return runKubectlNetworkPolicies(remaining)
		case "poddisruptionbudgets", "poddisruptionbudget", "pdb":
			return runKubectlPDB(remaining)
		case "customresourcedefinitions", "customresourcedefinition", "crd", "crds":
			return runKubectlCRDs(remaining)
		}
	}

	if args[0] == "logs" && len(args) > 1 {
		return runKubectlLogs(args[1:])
	}

	if args[0] == "describe" && len(args) > 1 {
		return runKubectlDescribe(args[1:])
	}

	if args[0] == "top" && len(args) > 1 {
		return runKubectlTop(args[1:])
	}

	if args[0] == "apply" || args[0] == "delete" {
		return runKubectlDryRunFirst(args)
	}

	return runKubectlPassthrough(args)
}

func runKubectlPassthrough(args []string) error {
	timer := tracking.Start()

	c := exec.Command("kubectl", args...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	if kubectlJSON {
		out.Global().Println(formatAsJSONKubectl(output))
		originalTokens := filter.EstimateTokens(output)
		timer.Track(fmt.Sprintf("kubectl %s", strings.Join(args, " ")), "tok kubectl", originalTokens, originalTokens)
		return err
	}

	out.Global().Print(output)

	originalTokens := filter.EstimateTokens(output)
	timer.Track(fmt.Sprintf("kubectl %s", strings.Join(args, " ")), "tok kubectl", originalTokens, originalTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runKubectlPods(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "pods", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	raw := stdoutBuf.String()

	if err != nil {
		out.Global().Print(stderrBuf.String())
		return err
	}

	var podList K8sPodList
	if err := unmarshalJSON(raw, &podList); err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		out.Global().Println("No pods found")
		timer.Track("kubectl get pods", "tok kubectl pods", filter.EstimateTokens(raw), 0)
		return nil
	}

	pods := podList.Items
	running, pending, failed, restartsTotal := 0, 0, 0, 0
	var issues []string

	for _, pod := range pods {
		ns := pod.Metadata.Namespace
		name := pod.Metadata.Name
		phase := pod.Status.Phase

		for _, cs := range pod.Status.ContainerStatuses {
			restartsTotal += cs.RestartCount
		}

		switch phase {
		case "Running":
			running++
		case "Pending":
			pending++
			issues = append(issues, fmt.Sprintf("%s/%s Pending", ns, name))
		case "Failed", "Error":
			failed++
			issues = append(issues, fmt.Sprintf("%s/%s %s", ns, name, phase))
		}
	}

	var parts []string
	if running > 0 {
		parts = append(parts, fmt.Sprintf("%d ok", running))
	}
	if pending > 0 {
		parts = append(parts, fmt.Sprintf("%d pending", pending))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%d failed", failed))
	}
	if restartsTotal > 0 {
		parts = append(parts, fmt.Sprintf("%d restarts", restartsTotal))
	}

	if shared.UltraCompact {
		filtered := fmt.Sprintf("%d pods: %s", len(pods), strings.Join(parts, ", "))
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("kubectl get pods", "tok kubectl pods", originalTokens, filteredTokens)
		return nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d pods: %s\n", len(pods), strings.Join(parts, ", ")))

	if len(issues) > 0 {
		result.WriteString("Issues:\n")
		for i, issue := range issues {
			if i >= 10 {
				break
			}
			result.WriteString(fmt.Sprintf("  %s\n", issue))
		}
		if len(issues) > 10 {
			result.WriteString(fmt.Sprintf("  ... +%d more\n", len(issues)-10))
		}
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get pods", "tok kubectl pods", originalTokens, filteredTokens)

	return nil
}

func runKubectlServices(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "services", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	raw := stdoutBuf.String()

	if err != nil {
		out.Global().Print(stderrBuf.String())
		return err
	}

	var svcList K8sServiceList
	if err := unmarshalJSON(raw, &svcList); err != nil {
		return err
	}
	if len(svcList.Items) == 0 {
		out.Global().Println("No services found")
		timer.Track("kubectl get services", "tok kubectl services", filter.EstimateTokens(raw), 0)
		return nil
	}

	services := svcList.Items

	if shared.UltraCompact {
		filtered := fmt.Sprintf("%d services:", len(services))
		for i, svc := range services {
			if i >= 5 {
				break
			}
			filtered += " " + svc.Metadata.Name
		}
		if len(services) > 5 {
			filtered += fmt.Sprintf(" +%d", len(services)-5)
		}
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("kubectl get services", "tok kubectl services", originalTokens, filteredTokens)
		return nil
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d services:\n", len(services)))

	for i, svc := range services {
		if i >= 15 {
			break
		}
		ns := svc.Metadata.Namespace
		name := svc.Metadata.Name
		svcType := svc.Spec.Type
		if svcType == "" {
			svcType = "ClusterIP"
		}

		var ports []string
		for _, p := range svc.Spec.Ports {
			port := p.Port
			target := p.TargetPort
			if target == "" || target == fmt.Sprintf("%d", port) {
				ports = append(ports, fmt.Sprintf("%d", port))
			} else {
				ports = append(ports, fmt.Sprintf("%d->%s", port, target))
			}
		}

		result.WriteString(fmt.Sprintf("  %s/%s %s [%s]\n", ns, name, svcType, strings.Join(ports, ",")))
	}

	if len(services) > 15 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", len(services)-15))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get services", "tok kubectl services", originalTokens, filteredTokens)

	return nil
}

func runKubectlDeployments(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "deployments", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var deployList K8sDeploymentList
	if err := unmarshalJSON(raw, &deployList); err != nil {
		return err
	}
	if len(deployList.Items) == 0 {
		out.Global().Println("No deployments found")
		timer.Track("kubectl get deployments", "tok kubectl deployments", filter.EstimateTokens(raw), 0)
		return nil
	}

	deployments := deployList.Items

	if shared.UltraCompact {
		filtered := fmt.Sprintf("%d deploys:", len(deployments))
		for i, d := range deployments {
			if i >= 5 {
				break
			}
			ready := d.Status.ReadyReplicas
			desired := d.Status.Replicas
			filtered += fmt.Sprintf(" %s(%d/%d)", d.Metadata.Name, ready, desired)
		}
		if len(deployments) > 5 {
			filtered += fmt.Sprintf(" +%d", len(deployments)-5)
		}
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("kubectl get deployments", "tok kubectl deployments", originalTokens, filteredTokens)
		return err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d deployments:\n", len(deployments)))

	for i, d := range deployments {
		if i >= 15 {
			break
		}
		name := d.Metadata.Name
		ns := d.Metadata.Namespace
		ready := d.Status.ReadyReplicas
		desired := d.Status.Replicas
		upToDate := d.Status.UpdatedReplicas
		available := d.Status.AvailableReplicas

		status := "ok"
		if ready != desired && desired > 0 {
			status = "updating"
		}
		if ready == 0 && desired > 0 {
			status = "down"
		}

		result.WriteString(fmt.Sprintf("  %s/%s %d/%d ready %d up %d avail [%s]\n",
			ns, name, ready, desired, upToDate, available, status))
	}

	if len(deployments) > 15 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", len(deployments)-15))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get deployments", "tok kubectl deployments", originalTokens, filteredTokens)

	return err
}

func runKubectlNodes(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "nodes", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var nodeList K8sNodeList
	if err := unmarshalJSON(raw, &nodeList); err != nil {
		return err
	}
	if len(nodeList.Items) == 0 {
		out.Global().Println("No nodes found")
		timer.Track("kubectl get nodes", "tok kubectl nodes", filter.EstimateTokens(raw), 0)
		return nil
	}

	nodes := nodeList.Items

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d nodes:\n", len(nodes)))

	for i, n := range nodes {
		if i >= 20 {
			break
		}
		name := n.Metadata.Name
		status := "Ready"
		for _, cond := range n.Status.Conditions {
			if cond.Type == "Ready" {
				status = cond.Status
				if status == "True" {
					status = "Ready"
				}
			}
		}

		cri := ""
		kubelet := ""
		os := ""
		arch := ""
		for k, v := range n.Status.NodeInfo {
			switch k {
			case "containerRuntimeVersion":
				cri = v
			case "kubeletVersion":
				kubelet = v
			case "operatingSystem":
				os = v
			case "architecture":
				arch = v
			}
		}

		capacity := ""
		for k, v := range n.Status.Capacity {
			if k == "cpu" {
				capacity = v
			}
		}

		if shared.UltraCompact {
			result.WriteString(fmt.Sprintf("  %s %s k8s=%s\n", name, status, kubelet))
		} else {
			result.WriteString(fmt.Sprintf("  %s [%s] k8s=%s cri=%s os=%s arch=%s cpu=%s\n",
				name, status, kubelet, cri, os, arch, capacity))
		}
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get nodes", "tok kubectl nodes", originalTokens, filteredTokens)

	return err
}

func runKubectlNamespaces(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "namespaces", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var nsList K8sNamespaceList
	if err := unmarshalJSON(raw, &nsList); err != nil {
		return err
	}
	if len(nsList.Items) == 0 {
		out.Global().Println("No namespaces found")
		timer.Track("kubectl get namespaces", "tok kubectl namespaces", filter.EstimateTokens(raw), 0)
		return nil
	}

	namespaces := nsList.Items

	if shared.UltraCompact {
		filtered := fmt.Sprintf("%d ns:", len(namespaces))
		for i, ns := range namespaces {
			if i >= 10 {
				break
			}
			filtered += " " + ns.Metadata.Name
		}
		if len(namespaces) > 10 {
			filtered += fmt.Sprintf(" +%d", len(namespaces)-10)
		}
		out.Global().Println(filtered)
		originalTokens := filter.EstimateTokens(raw)
		filteredTokens := filter.EstimateTokens(filtered)
		timer.Track("kubectl get namespaces", "tok kubectl namespaces", originalTokens, filteredTokens)
		return err
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d namespaces:\n", len(namespaces)))

	for i, ns := range namespaces {
		if i >= 20 {
			break
		}
		name := ns.Metadata.Name
		status := ns.Status.Phase
		result.WriteString(fmt.Sprintf("  %-30s %s\n", name, status))
	}

	if len(namespaces) > 20 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", len(namespaces)-20))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get namespaces", "tok kubectl namespaces", originalTokens, filteredTokens)

	return err
}

func runKubectlConfigMaps(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "configmaps", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var cmList K8sGenericList
	if err := unmarshalJSON(raw, &cmList); err != nil {
		return err
	}
	if len(cmList.Items) == 0 {
		out.Global().Println("No configmaps found")
		timer.Track("kubectl get configmaps", "tok kubectl configmaps", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("configmaps", cmList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get configmaps", "tok kubectl configmaps", originalTokens, filteredTokens)

	return err
}

func runKubectlSecrets(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "secrets", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var secretList K8sGenericList
	if err := unmarshalJSON(raw, &secretList); err != nil {
		return err
	}
	if len(secretList.Items) == 0 {
		out.Global().Println("No secrets found")
		timer.Track("kubectl get secrets", "tok kubectl secrets", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("secrets", secretList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get secrets", "tok kubectl secrets", originalTokens, filteredTokens)

	return err
}

func runKubectlIngress(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "ingress", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var ingressList K8sIngressList
	if err := unmarshalJSON(raw, &ingressList); err != nil {
		return err
	}
	if len(ingressList.Items) == 0 {
		out.Global().Println("No ingress found")
		timer.Track("kubectl get ingress", "tok kubectl ingress", filter.EstimateTokens(raw), 0)
		return nil
	}

	ingresses := ingressList.Items

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d ingress rules:\n", len(ingresses)))

	for i, ing := range ingresses {
		if i >= 15 {
			break
		}
		name := ing.Metadata.Name
		ns := ing.Metadata.Namespace

		var hosts []string
		var paths []string
		for _, rule := range ing.Spec.Rules {
			if rule.Host != "" {
				hosts = append(hosts, rule.Host)
			}
			for _, path := range rule.HTTP.Paths {
				paths = append(paths, fmt.Sprintf("%s->%s", path.Path, path.Backend.Service.Name))
			}
		}

		hostStr := strings.Join(hosts, ", ")
		if hostStr == "" {
			hostStr = "*"
		}
		pathStr := strings.Join(paths, ", ")
		if pathStr == "" {
			pathStr = "default"
		}

		result.WriteString(fmt.Sprintf("  %s/%s %s [%s]\n", ns, name, hostStr, pathStr))
	}

	if len(ingresses) > 15 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", len(ingresses)-15))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get ingress", "tok kubectl ingress", originalTokens, filteredTokens)

	return err
}

func runKubectlEvents(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "events", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	output, err := c.CombinedOutput()
	raw := string(output)

	var eventList K8sEventList
	if err := unmarshalJSON(raw, &eventList); err != nil {
		return err
	}
	if len(eventList.Items) == 0 {
		out.Global().Println("No events found")
		timer.Track("kubectl get events", "tok kubectl events", filter.EstimateTokens(raw), 0)
		return nil
	}

	events := eventList.Items

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%d events:\n", len(events)))

	for i, ev := range events {
		if i >= 20 {
			break
		}
		msg := ev.Message
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		reason := ev.Reason
		if reason == "" {
			reason = ev.Type
		}
		result.WriteString(fmt.Sprintf("  [%s] %s: %s\n", reason, ev.Metadata.Name, msg))
	}

	if len(events) > 20 {
		result.WriteString(fmt.Sprintf("  ... +%d more\n", len(events)-20))
	}

	filtered := result.String()
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get events", "tok kubectl events", originalTokens, filteredTokens)

	return err
}

func runKubectlLogs(args []string) error {
	timer := tracking.Start()

	if len(args) == 0 {
		out.Global().Println("Usage: tok kubectl logs <pod>")
		return nil
	}

	pod := args[0]
	kargs := []string{"logs", "--tail", "100", pod}
	kargs = append(kargs, args[1:]...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterLogOutput(output)
	out.Global().Printf("Logs for %s:\n%s", pod, filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("kubectl logs %s", pod), "tok kubectl logs", originalTokens, filteredTokens)

	return err
}

func runKubectlDescribe(args []string) error {
	timer := tracking.Start()

	c := exec.Command("kubectl", append([]string{"describe"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterKubectlDescribe(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("kubectl describe %s", strings.Join(args, " ")), "tok kubectl describe", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runKubectlTop(args []string) error {
	timer := tracking.Start()

	c := exec.Command("kubectl", append([]string{"top"}, args...)...)
	c.Env = os.Environ()

	stdoutBuf := shared.GetBuffer()
	defer shared.PutBuffer(stdoutBuf)
	stderrBuf := shared.GetBuffer()
	defer shared.PutBuffer(stderrBuf)
	c.Stdout = stdoutBuf
	c.Stderr = stderrBuf

	err := c.Run()
	output := stdoutBuf.String() + stderrBuf.String()

	filtered := filterKubectlTop(output)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(output)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("kubectl top %s", strings.Join(args, " ")), "tok kubectl top", originalTokens, filteredTokens)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

func runKubectlDryRunFirst(args []string) error {
	return runKubectlPassthrough(args)
}

func runKubectlJobs(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "jobs", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	var jobList K8sGenericList
	if err := unmarshalJSON(raw, &jobList); err != nil {
		return err
	}
	if len(jobList.Items) == 0 {
		out.Global().Println("No jobs found")
		timer.Track("kubectl get jobs", "tok kubectl jobs", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("jobs", jobList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get jobs", "tok kubectl jobs", originalTokens, filteredTokens)

	return err
}

func runKubectlCronJobs(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "cronjobs", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	var cjList K8sGenericList
	if err := unmarshalJSON(raw, &cjList); err != nil {
		return err
	}
	if len(cjList.Items) == 0 {
		out.Global().Println("No cronjobs found")
		timer.Track("kubectl get cronjobs", "tok kubectl cronjobs", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("cronjobs", cjList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get cronjobs", "tok kubectl cronjobs", originalTokens, filteredTokens)

	return err
}

func runKubectlPVs(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "persistentvolumes", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	var pvList K8sGenericList
	if err := unmarshalJSON(raw, &pvList); err != nil {
		return err
	}
	if len(pvList.Items) == 0 {
		out.Global().Println("No PVs found")
		timer.Track("kubectl get pv", "tok kubectl pv", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("persistentvolumes", pvList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get pv", "tok kubectl pv", originalTokens, filteredTokens)

	return err
}

func runKubectlPVCs(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "persistentvolumeclaims", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	var pvcList K8sGenericList
	if err := unmarshalJSON(raw, &pvcList); err != nil {
		return err
	}
	if len(pvcList.Items) == 0 {
		out.Global().Println("No PVCs found")
		timer.Track("kubectl get pvc", "tok kubectl pvc", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("persistentvolumeclaims", pvcList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get pvc", "tok kubectl pvc", originalTokens, filteredTokens)

	return err
}

func runKubectlServiceAccounts(args []string) error {
	return runKubectlGenericGet("serviceaccounts", "sa", args)
}

func runKubectlReplicaSets(args []string) error {
	return runKubectlGenericGet("replicasets", "rs", args)
}

func runKubectlDaemonSets(args []string) error {
	return runKubectlGenericGet("daemonsets", "ds", args)
}

func runKubectlStatefulSets(args []string) error {
	return runKubectlGenericGet("statefulsets", "sts", args)
}

func runKubectlHPA(args []string) error {
	return runKubectlGenericGet("horizontalpodautoscalers", "hpa", args)
}

func runKubectlNetworkPolicies(args []string) error {
	return runKubectlGenericGet("networkpolicies", "netpol", args)
}

func runKubectlPDB(args []string) error {
	return runKubectlGenericGet("poddisruptionbudgets", "pdb", args)
}

func runKubectlCRDs(args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", "crds", "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	var crdList K8sGenericList
	if err := unmarshalJSON(raw, &crdList); err != nil {
		return err
	}
	if len(crdList.Items) == 0 {
		out.Global().Println("No CRDs found")
		timer.Track("kubectl get crds", "tok kubectl crds", filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList("crds", crdList.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track("kubectl get crds", "tok kubectl crds", originalTokens, filteredTokens)

	return err
}

func runKubectlGenericGet(resource, shortName string, args []string) error {
	timer := tracking.Start()

	kargs := []string{"get", resource, "-o", "json"}
	kargs = append(kargs, args...)

	c := exec.Command("kubectl", kargs...)
	c.Env = os.Environ()
	output, err := c.CombinedOutput()
	raw := string(output)

	var list K8sGenericList
	if err := unmarshalJSON(raw, &list); err != nil {
		return err
	}
	if len(list.Items) == 0 {
		out.Global().Printf("No %s found\n", resource)
		timer.Track(fmt.Sprintf("kubectl get %s", resource), fmt.Sprintf("tok kubectl %s", shortName), filter.EstimateTokens(raw), 0)
		return nil
	}

	filtered := compactGenericList(resource, list.Items)
	out.Global().Print(filtered)

	originalTokens := filter.EstimateTokens(raw)
	filteredTokens := filter.EstimateTokens(filtered)
	timer.Track(fmt.Sprintf("kubectl get %s", resource), fmt.Sprintf("tok kubectl %s", shortName), originalTokens, filteredTokens)

	return err
}

func compactGenericList(resourceType string, items []K8sGenericItem) string {
	if shared.UltraCompact {
		result := fmt.Sprintf("%d %s:", len(items), resourceType)
		for i, item := range items {
			if i >= 5 {
				break
			}
			result += " " + item.Metadata.Name
		}
		if len(items) > 5 {
			result += fmt.Sprintf(" +%d", len(items)-5)
		}
		return result + "\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d %s:\n", len(items), resourceType))
	for i, item := range items {
		if i >= 20 {
			break
		}
		ns := item.Metadata.Namespace
		name := item.Metadata.Name
		if ns != "" && ns != "default" {
			sb.WriteString(fmt.Sprintf("  %s/%s\n", ns, name))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", name))
		}
	}
	if len(items) > 20 {
		sb.WriteString(fmt.Sprintf("  ... +%d more\n", len(items)-20))
	}
	return sb.String()
}

func filterKubectlDescribe(output string) string {
	var result strings.Builder
	importantKeys := map[string]bool{
		"Name:":          true,
		"Namespace:":     true,
		"Status:":        true,
		"Node:":          true,
		"Start":          true,
		"Containers:":    true,
		"Image:":         true,
		"Port:":          true,
		"State:":         true,
		"Last State:":    true,
		"Restart Count:": true,
		"IP:":            true,
		"Controlled By:": true,
		"Labels:":        true,
		"Annotations:":   true,
		"Conditions:":    true,
		"Type":           true,
		"Events:":        true,
		"Warning":        true,
		"Normal":         true,
		"Replicas:":      true,
		"Volumes:":       true,
		"Capacity:":      true,
		"Access Modes:":  true,
		"Storage Class:": true,
		"Selector:":      true,
		"Endpoints:":     true,
		"Session":        true,
	}

	inEvents := false
	eventCount := 0

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "Events:") || strings.HasPrefix(line, "  Type") {
			inEvents = true
			continue
		}

		if inEvents {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || (!strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "\t")) {
				inEvents = false
			} else {
				eventCount++
				if eventCount <= 10 {
					result.WriteString(shared.TruncateLine(line, 100) + "\n")
				}
				continue
			}
		}

		for key := range importantKeys {
			if strings.HasPrefix(line, "  "+key) || strings.HasPrefix(line, key) {
				truncated := shared.TruncateLine(line, 100)
				result.WriteString(truncated + "\n")
				break
			}
		}
	}

	if eventCount > 10 {
		result.WriteString(fmt.Sprintf("  ... +%d more events\n", eventCount-10))
	}

	if result.Len() == 0 {
		return filterLogOutput(output)
	}
	return result.String()
}

func filterKubectlTop(output string) string {
	var result strings.Builder
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		if i == 0 {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			name := fields[0]
			cpu := fields[1]
			mem := fields[2]
			if shared.UltraCompact {
				result.WriteString(fmt.Sprintf("%s cpu=%s mem=%s\n", name, cpu, mem))
			} else {
				result.WriteString(fmt.Sprintf("%-40s cpu=%-10s mem=%s\n", name, cpu, mem))
			}
		}
	}

	if result.Len() == 0 {
		return output
	}
	return result.String()
}

// JSON structures for kubectl

type K8sPodList struct {
	Items []K8sPod `json:"items"`
}

type K8sPod struct {
	Metadata K8sMetadata  `json:"metadata"`
	Status   K8sPodStatus `json:"status"`
}

type K8sMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type K8sPodStatus struct {
	Phase             string               `json:"phase"`
	ContainerStatuses []K8sContainerStatus `json:"containerStatuses"`
}

type K8sContainerStatus struct {
	Name         string         `json:"name"`
	RestartCount int            `json:"restartCount"`
	State        map[string]any `json:"state"`
}

type K8sServiceList struct {
	Items []K8sService `json:"items"`
}

type K8sService struct {
	Metadata K8sMetadata    `json:"metadata"`
	Spec     K8sServiceSpec `json:"spec"`
}

type K8sServiceSpec struct {
	Type  string           `json:"type"`
	Ports []K8sServicePort `json:"ports"`
}

type K8sServicePort struct {
	Port       int    `json:"port"`
	TargetPort string `json:"targetPort"`
}

type K8sDeploymentList struct {
	Items []K8sDeployment `json:"items"`
}

type K8sDeployment struct {
	Metadata K8sMetadata         `json:"metadata"`
	Spec     K8sDeploymentSpec   `json:"spec"`
	Status   K8sDeploymentStatus `json:"status"`
}

type K8sDeploymentSpec struct {
	Replicas int `json:"replicas"`
}

type K8sDeploymentStatus struct {
	Replicas          int `json:"replicas"`
	ReadyReplicas     int `json:"readyReplicas"`
	UpdatedReplicas   int `json:"updatedReplicas"`
	AvailableReplicas int `json:"availableReplicas"`
}

type K8sNodeList struct {
	Items []K8sNode `json:"items"`
}

type K8sNode struct {
	Metadata K8sMetadata   `json:"metadata"`
	Status   K8sNodeStatus `json:"status"`
}

type K8sNodeStatus struct {
	Conditions []K8sNodeCondition `json:"conditions"`
	NodeInfo   map[string]string  `json:"nodeInfo"`
	Capacity   map[string]string  `json:"capacity"`
}

type K8sNodeCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type K8sNamespaceList struct {
	Items []K8sNamespace `json:"items"`
}

type K8sNamespace struct {
	Metadata K8sMetadata        `json:"metadata"`
	Status   K8sNamespaceStatus `json:"status"`
}

type K8sNamespaceStatus struct {
	Phase string `json:"phase"`
}

type K8sGenericList struct {
	Items []K8sGenericItem `json:"items"`
}

type K8sGenericItem struct {
	Metadata K8sMetadata `json:"metadata"`
}

type K8sIngressList struct {
	Items []K8sIngress `json:"items"`
}

type K8sIngress struct {
	Metadata K8sMetadata    `json:"metadata"`
	Spec     K8sIngressSpec `json:"spec"`
}

type K8sIngressSpec struct {
	Rules []K8sIngressRule `json:"rules"`
}

type K8sIngressRule struct {
	Host string          `json:"host"`
	HTTP *K8sIngressHTTP `json:"http"`
}

type K8sIngressHTTP struct {
	Paths []K8sIngressPath `json:"paths"`
}

type K8sIngressPath struct {
	Path    string            `json:"path"`
	Backend K8sIngressBackend `json:"backend"`
}

type K8sIngressBackend struct {
	Service K8sIngressService `json:"service"`
}

type K8sIngressService struct {
	Name string `json:"name"`
}

type K8sEventList struct {
	Items []K8sEvent `json:"items"`
}

type K8sEvent struct {
	Metadata K8sMetadata `json:"metadata"`
	Type     string      `json:"type"`
	Reason   string      `json:"reason"`
	Message  string      `json:"message"`
}

func unmarshalJSON(data string, v any) error {
	return json.Unmarshal([]byte(data), v)
}

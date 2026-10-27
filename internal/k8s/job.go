package k8s

import (
	"fmt"
	"strings"

	"github.com/ksatchit/litmus-lite/internal/scenario"
)

// dnsLabel turns a scenario name into a name Kubernetes accepts and that is
// safe to write unquoted into YAML: lowercase letters, digits and '-', starting
// and ending with a letter or digit, short enough to take the "-scenario"
// suffix of the ConfigMap name (63 characters in all).
func dnsLabel(name string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 50 {
		out = strings.Trim(out[:50], "-")
	}
	if out == "" {
		return "litmus-lite"
	}
	return out
}

// JobYAML renders a Kubernetes Job that runs litmus-lite (no CRD, no ChaosEngine).
// The Job does not pass -yes: an unattended run keeps the guards that ask for
// it (disk.fill, process.kill, faults over a minute), and says so when one
// of them stops it.
func JobYAML(s *scenario.Scenario) string {
	name := dnsLabel(s.Metadata.Name)
	return fmt.Sprintf(`apiVersion: batch/v1
kind: Job
metadata:
  name: %s
spec:
  backoffLimit: 0
  template:
    spec:
      restartPolicy: Never
      containers:
        - name: litmus-lite
          image: ghcr.io/litmuschaos/litmus-lite:latest
          args: ["run", "/scenario.chaos.yaml"]
          volumeMounts:
            - name: scenario
              mountPath: /scenario.chaos.yaml
              subPath: scenario.chaos.yaml
      volumes:
        - name: scenario
          configMap:
            name: %s-scenario
`, name, name)
}

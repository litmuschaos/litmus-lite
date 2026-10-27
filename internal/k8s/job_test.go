package k8s

import (
	"strings"
	"testing"

	"github.com/ksatchit/litmus-lite/internal/scenario"
)

func TestJobYAML(t *testing.T) {
	s, err := scenario.Parse([]byte(`
apiVersion: litmus-lite.io/v1
kind: Scenario
metadata: { name: Demo_App }
faults:
  - name: a
    kind: http.latency
    duration: 1s
    params: { listen: "127.0.0.1:1", upstream: "127.0.0.1:2" }
probes:
  - name: p
    type: http
    url: http://127.0.0.1:1/
`))
	if err != nil {
		t.Fatal(err)
	}
	y := JobYAML(s)
	if !strings.Contains(y, "kind: Job") || strings.Contains(y, "ChaosEngine") {
		t.Fatal(y)
	}
}

func TestJobNameIsSafeInYAML(t *testing.T) {
	for in, want := range map[string]string{
		"Demo_App":                 "demo-app",
		"":                         "litmus-lite",
		"---":                      "litmus-lite",
		"a: b\nkind: Evil":         "a-b-kind-evil",
		`x"; rm -rf /`:             "x-rm-rf",
		strings.Repeat("abc-", 30): strings.Trim(strings.Repeat("abc-", 30)[:50], "-"),
	} {
		got := dnsLabel(in)
		if got != want {
			t.Errorf("dnsLabel(%q) = %q, want %q", in, got, want)
		}
		if len(got)+len("-scenario") > 63 {
			t.Errorf("%q: %q is too long for the ConfigMap name", in, got)
		}
	}
	s, err := scenario.Parse([]byte("apiVersion: litmus-lite.io/v1\nkind: Scenario\nmetadata:\n  name: \"evil\\nkind: Secret\"\nfaults:\n  - name: a\n    kind: http.latency\n    duration: 1s\n    params: { listen: \"127.0.0.1:1\", upstream: \"127.0.0.1:2\" }\nprobes:\n  - name: p\n    type: http\n    url: http://127.0.0.1:1/\n"))
	if err != nil {
		t.Fatal(err)
	}
	if y := JobYAML(s); strings.Count(y, "\nkind:") != 1 || strings.Contains(y, "Secret") && !strings.Contains(y, "evil-kind-secret") {
		t.Fatalf("a scenario name changed the manifest:\n%s", y)
	}
}

func TestJobDoesNotPassYes(t *testing.T) {
	s, _ := scenario.Parse([]byte("apiVersion: litmus-lite.io/v1\nkind: Scenario\nmetadata: { name: a }\nfaults:\n  - name: a\n    kind: http.latency\n    duration: 1s\n    params: { listen: \"127.0.0.1:1\", upstream: \"127.0.0.1:2\" }\nprobes:\n  - name: p\n    type: http\n    url: http://127.0.0.1:1/\n"))
	if strings.Contains(JobYAML(s), "-yes") {
		t.Fatal("the generated Job turns the safety guards off")
	}
}

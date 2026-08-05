package ml

import (
	"regexp"
	"strings"
	"sync"

	"github.com/oni1997/gatewayx/internal/history"
)

var (
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|drop\s+table|1\s*=\s*1|'\)\s*--|or\s+1\s*=\s*1)`)
	xssPattern          = regexp.MustCompile(`(?i)(<script|javascript:|onerror\s*=|onload\s*=|alert\s*\(|document\.cookie)`)
	pathTraversalPattern = regexp.MustCompile(`\.\./|\.\.\\`)
	shellInjectionPattern = regexp.MustCompile(`(?i)(;rm\s|;cat\s|\|.*sh|` + "`" + `.*` + "`" + `|\$\(.*\))`)
	suspiciousAgents     = []string{"nikto", "sqlmap", "nmap", "masscan", "gobuster", "dirbuster", "wfuzz"}
)

type SecurityReport struct {
	Threats     []Threat        `json:"threats"`
	TotalAlerts int             `json:"total_alerts"`
	Summary     string          `json:"summary"`
}

type Threat struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Path     string `json:"path"`
	Count    int    `json:"count"`
	Detail   string `json:"detail"`
}

type SecurityScanner struct {
	mu           sync.Mutex
	threatCounts map[string]*threatCounter
}

type threatCounter struct {
	Type   string
	Source string
	Path   string
	Count  int
	Detail string
}

func NewSecurityScanner() *SecurityScanner {
	return &SecurityScanner{
		threatCounts: make(map[string]*threatCounter),
	}
}

func (s *SecurityScanner) Scan(entries []history.Entry) SecurityReport {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.threatCounts = make(map[string]*threatCounter)

	statusCounts := make(map[int]int)
	sourceIpCounts := make(map[string]int)

	for _, e := range entries {
		statusCounts[e.Status]++
		sourceIpCounts[e.RemoteAddr]++

		s.checkSQLInjection(e)
		s.checkXSS(e)
		s.checkPathTraversal(e)
		s.checkShellInjection(e)
		s.checkBruteForce(e, sourceIpCounts)
	}

	report := SecurityReport{}

	if count := statusCounts[401]; count > 10 {
		report.Threats = append(report.Threats, Threat{
			Type: "auth_failure", Severity: "medium",
			Count: count, Detail: "high number of 401 responses, possible credential stuffing",
		})
	}
	if count := statusCounts[403]; count > 20 {
		report.Threats = append(report.Threats, Threat{
			Type: "access_denied", Severity: "medium",
			Count: count, Detail: "high number of 403 responses, possible enumeration attempt",
		})
	}
	if count := statusCounts[429]; count > 5 {
		report.Threats = append(report.Threats, Threat{
			Type: "rate_limited", Severity: "low",
			Count: count, Detail: "rate limit triggered, possible DoS attempt",
		})
	}

	for _, tc := range s.threatCounts {
		severity := "low"
		if tc.Count > 5 {
			severity = "high"
		} else if tc.Count > 2 {
			severity = "medium"
		}
		report.Threats = append(report.Threats, Threat{
			Type: tc.Type, Severity: severity,
			Source: tc.Source, Path: tc.Path,
			Count: tc.Count, Detail: tc.Detail,
		})
	}

	for ip, count := range sourceIpCounts {
		if count > 100 {
			report.Threats = append(report.Threats, Threat{
				Type: "high_volume", Severity: "high",
				Source: ip, Count: count,
				Detail: "single IP generating excessive traffic",
			})
		}
	}

	report.TotalAlerts = len(report.Threats)

	if report.TotalAlerts == 0 {
		report.Summary = "No threats detected. Traffic appears normal."
	} else if report.TotalAlerts < 3 {
		report.Summary = "Minor anomalies detected. Review the flagged items."
	} else {
		report.Summary = "Multiple threats detected. Immediate review recommended."
	}

	return report
}

func (s *SecurityScanner) checkSQLInjection(e history.Entry) {
	if sqlInjectionPattern.MatchString(e.Path) || sqlInjectionPattern.MatchString(e.Host) {
		key := "sqli:" + e.RemoteAddr
		if tc, ok := s.threatCounts[key]; ok {
			tc.Count++
		} else {
			s.threatCounts[key] = &threatCounter{
				Type: "sql_injection", Source: e.RemoteAddr,
				Path: e.Path, Count: 1,
				Detail: "SQL injection pattern detected in request path",
			}
		}
	}
}

func (s *SecurityScanner) checkXSS(e history.Entry) {
	if xssPattern.MatchString(e.Path) {
		key := "xss:" + e.RemoteAddr
		if tc, ok := s.threatCounts[key]; ok {
			tc.Count++
		} else {
			s.threatCounts[key] = &threatCounter{
				Type: "xss", Source: e.RemoteAddr,
				Path: e.Path, Count: 1,
				Detail: "Cross-site scripting pattern detected in request path",
			}
		}
	}
}

func (s *SecurityScanner) checkPathTraversal(e history.Entry) {
	if pathTraversalPattern.MatchString(e.Path) {
		key := "traversal:" + e.RemoteAddr
		if tc, ok := s.threatCounts[key]; ok {
			tc.Count++
		} else {
			s.threatCounts[key] = &threatCounter{
				Type: "path_traversal", Source: e.RemoteAddr,
				Path: e.Path, Count: 1,
				Detail: "Path traversal attempt detected",
			}
		}
	}
}

func (s *SecurityScanner) checkShellInjection(e history.Entry) {
	if shellInjectionPattern.MatchString(e.Path) {
		key := "shell:" + e.RemoteAddr
		if tc, ok := s.threatCounts[key]; ok {
			tc.Count++
		} else {
			s.threatCounts[key] = &threatCounter{
				Type: "shell_injection", Source: e.RemoteAddr,
				Path: e.Path, Count: 1,
				Detail: "Shell command injection pattern detected",
			}
		}
	}
}

func (s *SecurityScanner) checkBruteForce(e history.Entry, sourceCounts map[string]int) {
	ua := e.Host
	for _, agent := range suspiciousAgents {
		if strings.Contains(strings.ToLower(ua), agent) {
			key := "scanner:" + e.RemoteAddr
			if tc, ok := s.threatCounts[key]; ok {
				tc.Count++
			} else {
				s.threatCounts[key] = &threatCounter{
					Type: "scanner", Source: e.RemoteAddr,
					Detail: "Known security scanner user-agent detected: " + agent,
					Count: 1,
				}
			}
			return
		}
	}
}

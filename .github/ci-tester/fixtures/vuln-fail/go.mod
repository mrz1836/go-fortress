module fixture-vuln-fail

go 1.24

// This module intentionally depends on a package with known vulnerabilities
// for testing vulnerability detection tools.
// DO NOT UPDATE: golang.org/x/text must stay at v0.3.0 — it has known CVEs
// that SEC-003 expects govulncheck to detect.  Bumping this version will
// cause the Guardian CI Tester to fail.
require golang.org/x/text v0.3.0

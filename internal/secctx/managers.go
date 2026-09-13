package secctx

import "strings"

type ManagerWatch struct {
	Name string
	CIK  string
}

var CuratedManagers = []ManagerWatch{
	{Name: "BlackRock"},
	{Name: "Vanguard"},
	{Name: "State Street"},
	{Name: "Fidelity"},
	{Name: "JPMorgan"},
	{Name: "Goldman Sachs"},
	{Name: "Morgan Stanley"},
}

func ResolveWatchlist(resolved map[string]string) []ManagerWatch {
	var out []ManagerWatch
	for _, m := range CuratedManagers {
		cik, ok := resolved[strings.ToLower(m.Name)]
		if !ok || strings.TrimSpace(cik) == "" {
			continue
		}
		if _, err := parseCIKStrict(cik); err != nil {
			continue
		}
		out = append(out, ManagerWatch{Name: m.Name, CIK: ParseCIK(cik)})
	}
	return out
}

func parseCIKStrict(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", errCIK
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return "", errCIK
		}
	}
	return ParseCIK(s), nil
}

type cikErr string

func (e cikErr) Error() string { return string(e) }

const errCIK cikErr = "unresolved cik"

package installer

import "strings"

// ResolveTools translates the -tools flag value into optional tool indices.
//
//	""    → no optional tools (required only)
//	"all" → all optional tools
//	"a,b" → optional tools whose Name matches
func ResolveTools(tools []Tool, flag string) []int {
	flag = strings.TrimSpace(flag)
	if flag == "" {
		return nil
	}

	var optional []int
	for i, t := range tools {
		if !t.Required {
			optional = append(optional, i)
		}
	}

	if flag == "all" {
		return optional
	}

	wanted := map[string]bool{}
	for _, name := range strings.Split(flag, ",") {
		wanted[strings.TrimSpace(name)] = true
	}

	var result []int
	for _, i := range optional {
		if wanted[tools[i].Name] {
			result = append(result, i)
		}
	}
	return result
}

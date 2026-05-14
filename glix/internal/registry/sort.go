package registry

import (
	"sort"
	"strings"
)

func sortMatches(m []Match, query string) {
	q := strings.ToLower(query)
	sort.SliceStable(m, func(i, j int) bool {
		ni, nj := strings.ToLower(m[i].Name), strings.ToLower(m[j].Name)
		if q != "" {
			pi, pj := strings.HasPrefix(ni, q), strings.HasPrefix(nj, q)
			if pi != pj {
				return pi
			}
		}
		return ni < nj
	})
}

func sortStringsAlpha(s []string) {
	sort.Strings(s)
}

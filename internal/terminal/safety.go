package terminal

import (
	"net/http"
	"strings"
)

// BrokerMutationCapability is always NONE. This process is a read-only observer.
func BrokerMutationCapability() string { return "NONE" }

func AllowHTTP(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func ForbiddenVerbs() []string {
	return []string{"BUY", "SELL", "CLOSE", "FLATTEN"}
}

func ContainsMutationVerb(s string) bool {
	u := strings.ToUpper(strings.TrimSpace(s))
	for _, v := range ForbiddenVerbs() {
		if u == v {
			return true
		}
	}
	return false
}

func rejectMutations(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if AllowHTTP(r.Method) {
			if r.Method == http.MethodOptions {
				w.Header().Set("Allow", "GET, HEAD, OPTIONS")
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "read-only terminal", http.StatusMethodNotAllowed)
	})
}

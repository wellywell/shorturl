package auth

import (
	"net/http"
	"net/netip"
)

type SubnetChecker struct {
	Trusted string
}

func (s SubnetChecker) Handle(next http.Handler) http.Handler {

	authenticate := func(w http.ResponseWriter, r *http.Request) {

		ipStr := r.Header.Get("X-Real-IP")

		if ipStr == "" {
			http.Error(w, "Not trusted network", http.StatusForbidden)
			return
		}
		if !checkCIDR(ipStr, s.Trusted) {
			http.Error(w, "Not trusted network", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)

	}
	return http.HandlerFunc(authenticate)
}

func checkCIDR(ip string, mask string) bool {
	network, err := netip.ParsePrefix(mask)
	if err != nil {
		return false
	}

	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return false
	}
	return network.Contains(addr)

}

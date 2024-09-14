package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkCIDR(t *testing.T) {
	type args struct {
		ip   string
		mask string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"true", args{"192.168.0.1", "192.168.0.0/24"}, true},
		{"true", args{"192.168.0.1", "192.168.1.0/24"}, false},
		{"bad", args{"192", "192.168.1.0/24"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkCIDR(tt.args.ip, tt.args.mask); got != tt.want {
				t.Errorf("checkCIDR() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MockHandler struct{}

func (MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func TestHandleWithIP(t *testing.T) {

	testCases := []struct {
		realIP     string
		resultCode int
	}{
		{"192.168.0.1", http.StatusOK},
		{"", http.StatusForbidden},
		{"123", http.StatusForbidden},
	}

	handler := &SubnetChecker{Trusted: "192.168.0.0/24"}
	mock := MockHandler{}

	for _, tc := range testCases {
		t.Run(tc.realIP, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.Header.Set("X-REAL-IP", tc.realIP)
			w := httptest.NewRecorder()

			testHandler := handler.Handle(mock)
			testHandler.ServeHTTP(w, r)

			assert.Equal(t, tc.resultCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

func TestHandleEmptyHeaderP(t *testing.T) {

	handler := &SubnetChecker{Trusted: "192.168.0.0/24"}
	mock := MockHandler{}

	t.Run("noheader", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		testHandler := handler.Handle(mock)
		testHandler.ServeHTTP(w, r)

		assert.Equal(t, http.StatusForbidden, w.Code, "Код ответа не совпадает с ожидаемым")
	})
}

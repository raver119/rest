package rest

import (
	"github.com/coredns/caddy"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_setup(t *testing.T) {

	tests := []struct {
		name        string
		config      string
		expectedUrl string
		expectedTtl uint32
		wantErr     bool
	}{
		{"test_0", "rest", "", 0, true},
		{"test_1", "rest bad_param", "", 0, true},
		{"test_2", "rest http://example.org/rest/ bad_number", "", 0, true},
		{"test_3", "rest http://example.org/rest/ 32", "http://example.org/rest/", 32, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := caddy.NewTestController("dns", tt.config)
			p, err := setupPlugin(c)

			assert.Equal(t, (err != nil), tt.wantErr, "setup() error = %v, wantErr %v", err, tt.wantErr )

			assert.Equal(t, p.url, tt.expectedUrl)
			assert.Equal(t, p.ttl, tt.expectedTtl)
		})
	}
}

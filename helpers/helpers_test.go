package helpers_test

import (
	"authex/helpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComputeMarketAddress(t *testing.T) {
	tests := []struct {
		name         string
		baseAddress  string
		quoteAddress string
		want         string
		wantErr      error
	}{
		{
			name:         "ERR: invalid input",
			baseAddress:  "abc",
			quoteAddress: "def",
			want:         "",
			wantErr:      helpers.ErrInput,
		},
		{
			name:         "ERR: invalid input 2",
			baseAddress:  "",
			quoteAddress: "0xbbD65e1115Ff895b6c0F313ca050A613a150c940",
			want:         "",
			wantErr:      helpers.ErrInput,
		},
		{
			name:         "OK",
			baseAddress:  "0xaa992902d88EA6192585B72D0B01C020F036bb99",
			quoteAddress: "0xbbD65e1115Ff895b6c0F313ca050A613a150c940",
			want:         "0x36f5e0ce0a49c8b10ae4e0d5214cda5d8b46073d",
			wantErr:      nil,
		},
		{
			name:         "OK: same result different order",
			baseAddress:  "0xbbD65e1115Ff895b6c0F313ca050A613a150c940",
			quoteAddress: "0xaa992902d88EA6192585B72D0B01C020F036bb99",
			want:         "0x36f5e0ce0a49c8b10ae4e0d5214cda5d8b46073d",
			wantErr:      nil,
		},
		{
			name:         "OK: same result different case",
			baseAddress:  "0xbbd65e1115ff895b6c0f313ca050a613a150c940",
			quoteAddress: "0xaa992902d88EA6192585B72D0B01C020F036bb99",
			want:         "0x36f5e0ce0a49c8b10ae4e0d5214cda5d8b46073d",
			wantErr:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := helpers.ComputeMarketAddress(tt.baseAddress, tt.quoteAddress)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, got, tt.want)
		})
	}
}

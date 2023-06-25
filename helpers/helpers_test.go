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
		{
			name:         "OK: usd/eur",
			baseAddress:  helpers.AsAddress("USD"),
			quoteAddress: helpers.AsAddress("EUR"),
			want:         "0xd36cfda1a6607e8b79d0c9ea784346a6e21fad86",
			wantErr:      nil,
		},
		{
			name:         "OK: eth/eur",
			baseAddress:  helpers.AsAddress("ETH"),
			quoteAddress: helpers.AsAddress("EUR"),
			want:         "0x98e08472d3cf60929829c4e252913d0295e64f33",
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

func TestAsAddress(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "OK",
			input: "USD",
			want:  "0x505f49beeda8b41a13274e3622c64e61d087a796",
		},
		{
			name:  "OK",
			input: "EUR",
			want:  "0x60c197cc20da7f7d7c4d019fb9e66cd79b223c6c",
		},
		{
			name:  "OK",
			input: "ETH",
			want:  "0x08db13fc7a9adf7ca72641f84d75b47069d3d7f0",
		},
		{
			name:  "OK",
			input: "eth",
			want:  "0x08db13fc7a9adf7ca72641f84d75b47069d3d7f0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, helpers.AsAddress(tt.input), tt.want)
		})
	}
}

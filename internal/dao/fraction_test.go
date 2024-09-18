package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_calculateRatio(t *testing.T) {
	type args struct {
		dPercents []delegatesForFraction
	}
	tests := []struct {
		name string
		args args
		want map[string]int
	}{
		{
			name: "Test 1",
			args: args{
				dPercents: []delegatesForFraction{
					{
						address: "0x1",
						percent: 25,
					},
					{
						address: "0x2",
						percent: 25,
					},
					{
						address: "0x3",
						percent: 50,
					},
				},
			},
			want: map[string]int{
				"0x1": 1,
				"0x2": 1,
				"0x3": 2,
			},
		},
		{
			name: "Test 2",
			args: args{
				dPercents: []delegatesForFraction{
					{
						address: "0x1",
						percent: 33,
					},
					{
						address: "0x2",
						percent: 33,
					},
					{
						address: "0x3",
						percent: 33,
					},
				},
			},
			want: map[string]int{
				"0x1": 1,
				"0x2": 1,
				"0x3": 1,
			},
		},
		{
			name: "Test 3",
			args: args{
				dPercents: []delegatesForFraction{
					{
						address: "0x1",
						percent: 28,
					},
					{
						address: "0x2",
						percent: 56,
					},
				},
			},
			want: map[string]int{
				"0x1": 1,
				"0x2": 2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, calculateRatio(tt.args.dPercents), "calculateRatio(%v)", tt.args.dPercents)
		})
	}
}

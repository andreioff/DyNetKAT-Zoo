package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert_NewFlowRule(t *testing.T) {
	cases := map[string]struct {
		inPort      int64
		outPort     int64
		isLink      bool
		assertSetup func(*testing.T, FlowRule)
	}{
		"Valid Flow Rule [Success]": {
			inPort:  4,
			outPort: -1,
			isLink:  true,
			assertSetup: func(t *testing.T, fr FlowRule) {
				assert.Equal(t, int64(4), fr.inPort)
				assert.Equal(t, int64(-1), fr.outPort)
				assert.True(t, fr.isLink)
			},
		},
		"Flow Rule Getters [Success]": {
			inPort:  -100,
			outPort: 16,
			isLink:  false,
			assertSetup: func(t *testing.T, fr FlowRule) {
				assert.Equal(t, int64(-100), fr.InPort())
				assert.Equal(t, int64(16), fr.OutPort())
				assert.False(t, fr.IsLink())
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fr := NewFlowRule(tc.inPort, tc.outPort, tc.isLink)
			// Assert the result
			if tc.assertSetup != nil {
				tc.assertSetup(t, fr)
			}
		})
	}
}

func TestFlowRule_IsEqual(t *testing.T) {
	fr2 := FlowRule{
		inPort:  0,
		outPort: 189,
		isLink:  true,
	}

	type fields struct {
		inPort  int64
		outPort int64
		isLink  bool
	}
	type args struct {
		fr2 FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "No matching fields [Success]",
			fields: fields{
				inPort:  1,
				outPort: 188,
				isLink:  false,
			},
			args: args{fr2: fr2},
			want: false,
		},
		{
			name: "One matching field [Success]",
			fields: fields{
				inPort:  0,
				outPort: 188,
				isLink:  false,
			},
			args: args{fr2: fr2},
			want: false,
		},
		{
			name: "Two matching fields [Success]",
			fields: fields{
				inPort:  0,
				outPort: 188,
				isLink:  true,
			},
			args: args{fr2: fr2},
			want: false,
		},
		{
			name: "All matching fields [Success]",
			fields: fields{
				inPort:  0,
				outPort: 189,
				isLink:  true,
			},
			args: args{fr2: fr2},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fr1 := FlowRule{
				inPort:  tt.fields.inPort,
				outPort: tt.fields.outPort,
				isLink:  tt.fields.isLink,
			}

			assert.Equal(t, tt.want, fr1.IsEqual(tt.args.fr2))
		})
	}
}

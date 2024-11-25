package convert

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

func getEmptyEntries() map[int64][]FlowRule {
	return make(map[int64][]FlowRule)
}

func getMockFTEntries1() map[int64][]FlowRule {
	return map[int64][]FlowRule{
		0: {{10, 11, false}, {10, 12, true}},
		1: {{13, 14, false}},
		3: {{15, 16, false}, {15, 17, true}},
	}
}

func getMockFTEntries2() map[int64][]FlowRule {
	return map[int64][]FlowRule{
		4: {{30, 31, false}, {30, 32, true}},
		6: {{33, 34, false}},
		2: {{35, 36, false}, {35, 37, true}},
	}
}

func getMockFTEntries3() map[int64][]FlowRule {
	return map[int64][]FlowRule{
		0: {{10, 11, false}, {10, 12, true}},
		1: {{13, 14, false}, {13, 19, false}, {13, 20, true}},
		3: {{15, 16, false}, {15, 17, true}},
	}
}

func getMockEmptyFT() *FlowTable {
	return &FlowTable{getEmptyEntries()}
}

func getMockFT1() *FlowTable {
	return &FlowTable{getMockFTEntries1()}
}

func getMockFT2() *FlowTable {
	return &FlowTable{getMockFTEntries2()}
}

func getMockFT3() *FlowTable {
	return &FlowTable{getMockFTEntries3()}
}

func TestNewFlowTable(t *testing.T) {
	tests := []struct {
		name string
		want *FlowTable
	}{
		{
			name: "Valid Flow Table [Success]",
			want: &FlowTable{
				entries: getEmptyEntries(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewFlowTable())
		})
	}
}

func TestFlowTable_Entries(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		want   map[int64][]FlowRule
	}{
		{
			name: "Get flow table entries (empty) [Success]",
			fields: fields{
				entries: getEmptyEntries(),
			},
			want: getEmptyEntries(),
		},
		{
			name: "Get flow table entries (non-empty) [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			want: getMockFTEntries1(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			assert.Equal(t, tt.want, ft.Entries())
		})
	}
}

func TestFlowTable_setEntries(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	type args struct {
		newEntries map[int64][]FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[int64][]FlowRule
	}{
		{
			name: "Set empty flow table entries [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				newEntries: getEmptyEntries(),
			},
			want: getEmptyEntries(),
		},
		{
			name: "Set flow table entries [Success]",
			fields: fields{
				entries: getEmptyEntries(),
			},
			args: args{
				newEntries: getMockFTEntries1(),
			},
			want: getMockFTEntries1(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			ft.setEntries(tt.args.newEntries)
			assert.Equal(t, tt.want, ft.Entries())
		})
	}
}

func TestFlowTable_hasEntry(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	type args struct {
		hostId int64
		fr     FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Empty Flow Table [Success]",
			fields: fields{
				entries: getEmptyEntries(),
			},
			args: args{
				hostId: -1,
				fr:     FlowRule{1, 45, false},
			},
			want: false,
		},
		{
			name: "No Matching Host Id [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				hostId: 2,
				fr:     FlowRule{10, 11, false},
			},
			want: false,
		},
		{
			name: "No Matching FlowRule [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				hostId: 0,
				fr:     FlowRule{10, 11, true},
			},
			want: false,
		},
		{
			name: "Matching Entry [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				hostId: 3,
				fr:     FlowRule{15, 16, false},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			assert.Equal(t, tt.want, ft.hasEntry(tt.args.hostId, tt.args.fr))
		})
	}
}

func TestFlowTable_AddEntry(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	type args struct {
		destHostId int64
		fr         FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FlowTable
	}{
		{
			name: "Existing Entry [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				destHostId: 1,
				fr:         FlowRule{13, 14, false},
			},
			want: getMockFT1(),
		},
		{
			name: "New Entry With Existing Host Id [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				destHostId: 1,
				fr:         FlowRule{15, 16, true},
			},
			want: &FlowTable{map[int64][]FlowRule{
				0: {{10, 11, false}, {10, 12, true}},
				1: {{13, 14, false}, {15, 16, true}},
				3: {{15, 16, false}, {15, 17, true}},
			}},
		},
		{
			name: "New Entry With New Host Id [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				destHostId: -3,
				fr:         FlowRule{18, 19, false},
			},
			want: &FlowTable{map[int64][]FlowRule{
				0:  {{10, 11, false}, {10, 12, true}},
				1:  {{13, 14, false}},
				3:  {{15, 16, false}, {15, 17, true}},
				-3: {{18, 19, false}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			ft.AddEntry(tt.args.destHostId, tt.args.fr)
			assert.EqualValues(t, tt.want, ft)
		})
	}
}

func TestFlowTable_Filter(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	type args struct {
		pred func(FlowRule) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FlowTable
	}{
		{
			name: "Always false predicate [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				pred: func(FlowRule) bool {
					return false
				},
			},
			want: getMockEmptyFT(),
		},
		{
			name: "Always true predicate [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				pred: func(FlowRule) bool {
					return true
				},
			},
			want: getMockFT1(),
		},
		{
			name: "Valid predicate [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				pred: func(ft FlowRule) bool {
					return ft.isLink
				},
			},
			want: &FlowTable{map[int64][]FlowRule{
				0: {{10, 12, true}},
				3: {{15, 17, true}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			got := ft.Filter(tt.args.pred)
			assert.Equal(t, tt.want, got)
			assert.NotSame(t, ft, got)
		})
	}
}

func TestFlowTable_Extend(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	type args struct {
		otherFt *FlowTable
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *FlowTable
	}{
		{
			name: "Nil argument [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				nil,
			},
			want: getMockFT1(),
		},
		{
			name: "Empty flow table argument [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				getMockEmptyFT(),
			},
			want: getMockFT1(),
		},
		{
			name: "Empty flow table base [Success]",
			fields: fields{
				entries: getEmptyEntries(),
			},
			args: args{
				getMockFT1(),
			},
			want: getMockFT1(),
		},
		{
			name: "No new entries [Success]",
			fields: fields{
				entries: getMockFTEntries3(),
			},
			args: args{
				getMockFT1(),
			},
			want: getMockFT3(),
		},
		{
			name: "New and existing entries [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			args: args{
				getMockFT3(),
			},
			want: getMockFT3(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			ft.Extend(tt.args.otherFt)
			assert.Equal(t, tt.want, ft)
		})
	}
}

func TestFlowTable_Copy(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		want   *FlowTable
	}{
		{
			name: "No entries [Success]",
			fields: fields{
				entries: getEmptyEntries(),
			},
			want: getMockEmptyFT(),
		},
		{
			name: "Non-empty flow table [Success]",
			fields: fields{
				entries: getMockFTEntries2(),
			},
			want: getMockFT2(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			got := ft.Copy()
			assert.Equal(t, tt.want, got)
			assert.NotSame(t, ft, got)
		})
	}
}

func TestFlowTable_ToNetKATPolicies(t *testing.T) {
	type fields struct {
		entries map[int64][]FlowRule
	}
	tests := []struct {
		name   string
		fields fields
		want   []*SimpleNetKATPolicy
	}{
		{
			name: "No Entries [Success]",
			fields: fields{
				entries: getEmptyEntries(),
			},
			want: []*SimpleNetKATPolicy{},
		},
		{
			name: "Non-emtpy Flow Table [Success]",
			fields: fields{
				entries: getMockFTEntries1(),
			},
			want: []*SimpleNetKATPolicy{
				{
					[]util.StrTup{{Fst: DST_STRING, Snd: "0"}, {Fst: PORT_STRING, Snd: "10"}},
					[]util.StrTup{{Fst: PORT_STRING, Snd: "11"}},
				},
				{
					[]util.StrTup{{Fst: DST_STRING, Snd: "0"}, {Fst: PORT_STRING, Snd: "10"}},
					[]util.StrTup{{Fst: PORT_STRING, Snd: "12"}},
				},
				{
					[]util.StrTup{{Fst: DST_STRING, Snd: "1"}, {Fst: PORT_STRING, Snd: "13"}},
					[]util.StrTup{{Fst: PORT_STRING, Snd: "14"}},
				},
				{
					[]util.StrTup{{Fst: DST_STRING, Snd: "3"}, {Fst: PORT_STRING, Snd: "15"}},
					[]util.StrTup{{Fst: PORT_STRING, Snd: "16"}},
				},
				{
					[]util.StrTup{{Fst: DST_STRING, Snd: "3"}, {Fst: PORT_STRING, Snd: "15"}},
					[]util.StrTup{{Fst: PORT_STRING, Snd: "17"}},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ft := &FlowTable{
				entries: tt.fields.entries,
			}
			assert.Equal(t, tt.want, ft.ToNetKATPolicies())
		})
	}
}

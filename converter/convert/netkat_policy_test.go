package convert

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

func TestNewSimpleNetKATPolicy(t *testing.T) {
	tests := []struct {
		name string
		want *SimpleNetKATPolicy
	}{
		{
			name: "Valid policy [Success]",
			want: &SimpleNetKATPolicy{
				completeTest:       []util.StrTup{},
				completeAssignment: []util.StrTup{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSimpleNetKATPolicy(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSimpleNetKATPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleNetKATPolicy_AddTest(t *testing.T) {
	type fields struct {
		completeTest       []util.StrTup
		completeAssignment []util.StrTup
	}
	type args struct {
		fieldName  string
		fieldValue string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		assertSetup func(*testing.T, *SimpleNetKATPolicy)
	}{
		{
			name: "Empty fields [Success]",
			fields: fields{
				completeTest:       []util.StrTup{},
				completeAssignment: []util.StrTup{},
			},
			args: args{
				fieldName:  "",
				fieldValue: "",
			},
			assertSetup: func(t *testing.T, snp *SimpleNetKATPolicy) {
				assert.Len(t, snp.completeTest, 1)
				assert.Len(t, snp.completeAssignment, 0)
				assert.EqualValues(t, snp.completeTest, []util.StrTup{{Fst: "", Snd: ""}})
			},
		},
		{
			name: "Duplicate fields and values [Success]",
			fields: fields{
				completeTest:       []util.StrTup{{Fst: "fieldTest", Snd: "valueTest"}},
				completeAssignment: []util.StrTup{{Fst: "fieldAssign", Snd: "valueAssign"}},
			},
			args: args{
				fieldName:  "fieldTest",
				fieldValue: "valueTest",
			},
			assertSetup: func(t *testing.T, snp *SimpleNetKATPolicy) {
				assert.Len(t, snp.completeTest, 2)
				assert.Len(t, snp.completeAssignment, 1)
				assert.EqualValues(
					t,
					snp.completeTest,
					[]util.StrTup{
						{Fst: "fieldTest", Snd: "valueTest"},
						{Fst: "fieldTest", Snd: "valueTest"},
					},
				)
				assert.EqualValues(
					t,
					snp.completeAssignment,
					[]util.StrTup{
						{Fst: "fieldAssign", Snd: "valueAssign"},
					},
				)
			},
		},
		{
			name: "Valid fields and values [Success]",
			fields: fields{
				completeTest:       []util.StrTup{{Fst: "fieldTest1", Snd: "valueTest1"}},
				completeAssignment: []util.StrTup{{Fst: "fieldAssign", Snd: "valueAssign"}},
			},
			args: args{
				fieldName:  "fieldTest2",
				fieldValue: "valueTest2",
			},
			assertSetup: func(t *testing.T, snp *SimpleNetKATPolicy) {
				assert.Len(t, snp.completeTest, 2)
				assert.Len(t, snp.completeAssignment, 1)
				assert.EqualValues(
					t,
					snp.completeTest,
					[]util.StrTup{
						{Fst: "fieldTest1", Snd: "valueTest1"},
						{Fst: "fieldTest2", Snd: "valueTest2"},
					},
				)
				assert.EqualValues(
					t,
					snp.completeAssignment,
					[]util.StrTup{
						{Fst: "fieldAssign", Snd: "valueAssign"},
					},
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snp := &SimpleNetKATPolicy{
				completeTest:       tt.fields.completeTest,
				completeAssignment: tt.fields.completeAssignment,
			}
			snp.AddTest(tt.args.fieldName, tt.args.fieldValue)
			// Assert the result
			if tt.assertSetup != nil {
				tt.assertSetup(t, snp)
			}
		})
	}
}

func TestSimpleNetKATPolicy_AddAssignment(t *testing.T) {
	type fields struct {
		completeTest       []util.StrTup
		completeAssignment []util.StrTup
	}
	type args struct {
		fieldName  string
		fieldValue string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		assertSetup func(*testing.T, *SimpleNetKATPolicy)
	}{
		{
			name: "Empty fields [Success]",
			fields: fields{
				completeTest:       []util.StrTup{},
				completeAssignment: []util.StrTup{},
			},
			args: args{
				fieldName:  "",
				fieldValue: "",
			},
			assertSetup: func(t *testing.T, snp *SimpleNetKATPolicy) {
				assert.Len(t, snp.completeTest, 0)
				assert.Len(t, snp.completeAssignment, 1)
				assert.EqualValues(t, snp.completeAssignment, []util.StrTup{{Fst: "", Snd: ""}})
			},
		},
		{
			name: "Duplicate fields and values [Success]",
			fields: fields{
				completeTest:       []util.StrTup{{Fst: "fieldTest", Snd: "valueTest"}},
				completeAssignment: []util.StrTup{{Fst: "fieldAssign", Snd: "valueAssign"}},
			},
			args: args{
				fieldName:  "fieldAssign",
				fieldValue: "valueAssign",
			},
			assertSetup: func(t *testing.T, snp *SimpleNetKATPolicy) {
				assert.Len(t, snp.completeTest, 1)
				assert.Len(t, snp.completeAssignment, 2)
				assert.EqualValues(
					t,
					snp.completeTest,
					[]util.StrTup{
						{Fst: "fieldTest", Snd: "valueTest"},
					},
				)
				assert.EqualValues(
					t,
					snp.completeAssignment,
					[]util.StrTup{
						{Fst: "fieldAssign", Snd: "valueAssign"},
						{Fst: "fieldAssign", Snd: "valueAssign"},
					},
				)
			},
		},
		{
			name: "Valid fields and values [Success]",
			fields: fields{
				completeTest:       []util.StrTup{{Fst: "fieldTest", Snd: "valueTest"}},
				completeAssignment: []util.StrTup{{Fst: "fieldAssign1", Snd: "valueAssign1"}},
			},
			args: args{
				fieldName:  "fieldAssign2",
				fieldValue: "valueAssign2",
			},
			assertSetup: func(t *testing.T, snp *SimpleNetKATPolicy) {
				assert.Len(t, snp.completeTest, 1)
				assert.Len(t, snp.completeAssignment, 2)
				assert.EqualValues(
					t,
					snp.completeTest,
					[]util.StrTup{
						{Fst: "fieldTest", Snd: "valueTest"},
					},
				)
				assert.EqualValues(
					t,
					snp.completeAssignment,
					[]util.StrTup{
						{Fst: "fieldAssign1", Snd: "valueAssign1"},
						{Fst: "fieldAssign2", Snd: "valueAssign2"},
					},
				)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snp := &SimpleNetKATPolicy{
				completeTest:       tt.fields.completeTest,
				completeAssignment: tt.fields.completeAssignment,
			}
			snp.AddAssignment(tt.args.fieldName, tt.args.fieldValue)
			// Assert the result
			if tt.assertSetup != nil {
				tt.assertSetup(t, snp)
			}
		})
	}
}

func TestSimpleNetKATPolicy_ToString(t *testing.T) {
	type fields struct {
		completeTest       []util.StrTup
		completeAssignment []util.StrTup
	}
	type args struct {
		AndSym    string
		EqSym     string
		AssignSym string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "No tests or assignments [Success]",
			fields: fields{
				completeTest:       []util.StrTup{},
				completeAssignment: []util.StrTup{},
			},
			args: args{
				AndSym:    "+",
				EqSym:     "=",
				AssignSym: "<-",
			},
			want: "",
		},
		{
			name: "One test [Success]",
			fields: fields{
				completeTest:       []util.StrTup{{Fst: "fieldTest1", Snd: "valueTest1"}},
				completeAssignment: []util.StrTup{},
			},
			args: args{
				AndSym:    "+",
				EqSym:     "=",
				AssignSym: "<-",
			},
			want: "(fieldTest1=valueTest1)",
		},
		{
			name: "One assignment [Success]",
			fields: fields{
				completeTest:       []util.StrTup{},
				completeAssignment: []util.StrTup{{Fst: "fieldAssign1", Snd: "valueAssign1"}},
			},
			args: args{
				AndSym:    "+",
				EqSym:     "=",
				AssignSym: "<-",
			},
			want: "(fieldAssign1<-valueAssign1)",
		},
		{
			name: "Tests and assignments [Success]",
			fields: fields{
				completeTest: []util.StrTup{
					{Fst: "fieldTest1", Snd: "valueTest1"},
					{Fst: "fieldTest2", Snd: "valueTest2"},
				},
				completeAssignment: []util.StrTup{
					{Fst: "fieldAssign1", Snd: "valueAssign1"},
					{Fst: "fieldAssign2", Snd: "valueAssign2"},
				},
			},
			args: args{
				AndSym:    "+",
				EqSym:     "=",
				AssignSym: "<-",
			},
			want: "(fieldTest1=valueTest1)+(fieldTest2=valueTest2)+(fieldAssign1<-valueAssign1)+(fieldAssign2<-valueAssign2)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snp := &SimpleNetKATPolicy{
				completeTest:       tt.fields.completeTest,
				completeAssignment: tt.fields.completeAssignment,
			}
			if got := snp.ToString(tt.args.AndSym, tt.args.EqSym, tt.args.AssignSym); got != tt.want {
				t.Errorf("SimpleNetKATPolicy.ToString() = %v, want %v", got, tt.want)
			}
		})
	}
}

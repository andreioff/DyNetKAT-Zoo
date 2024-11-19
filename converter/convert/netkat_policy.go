package convert

import (
	"fmt"
	"strings"

	"utwente.nl/topology-to-dynetkat-coverter/util"
)

/*
A simple NetKAT policy contains a complete test and a complete assignment for packet fields.
A complete test is a list of string tuples (field name, field value) specifying the values
that the fileds of a packet should have to pass the test.
Analogously, a complete assignment specifies the values that will be assigned to packet fields.
*/
type SimpleNetKATPolicy struct {
	completeTest       []util.StrTup
	completeAssignment []util.StrTup
}

func NewSimpleNetKATPolicy() *SimpleNetKATPolicy {
	return &SimpleNetKATPolicy{
		completeTest:       []util.StrTup{},
		completeAssignment: []util.StrTup{},
	}
}

func (snp *SimpleNetKATPolicy) AddTest(fieldName, fieldValue string) {
	snp.completeTest = append(snp.completeTest, util.NewStrTup(fieldName, fieldValue))
}

func (snp *SimpleNetKATPolicy) AddAssignment(fieldName, fieldValue string) {
	snp.completeAssignment = append(snp.completeAssignment, util.NewStrTup(fieldName, fieldValue))
}

func (snp *SimpleNetKATPolicy) ToString(AndSym, EqSym, AssignSym string) string {
	var sb strings.Builder

	prefix := ""
	for _, test := range snp.completeTest {
		sb.WriteString(fmt.Sprintf("%s(%s%s%s)", prefix, test.Fst, EqSym, test.Snd))
		prefix = AndSym
	}

	for _, assig := range snp.completeAssignment {
		sb.WriteString(fmt.Sprintf("%s(%s%s%s)", prefix, assig.Fst, AssignSym, assig.Snd))
	}

	return sb.String()
}

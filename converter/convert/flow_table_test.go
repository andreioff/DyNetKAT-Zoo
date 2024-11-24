package convert

func getMockFT1() *FlowTable {
	return &FlowTable{entries: map[int64][]FlowRule{
		0: {{10, 11, false}, {10, 12, true}},
		1: {{13, 14, false}},
		3: {{15, 16, false}, {15, 17, true}},
	}}
}

func getMockFT2() *FlowTable {
	return &FlowTable{
		entries: map[int64][]FlowRule{
			4: {{30, 31, false}, {30, 32, true}},
			6: {{33, 34, false}},
			2: {{35, 36, false}, {35, 37, true}},
		},
	}
}

func getMockFT3() *FlowTable {
	return &FlowTable{
		map[int64][]FlowRule{
			0: {{10, 11, false}, {10, 12, true}},
			1: {{13, 14, false}, {13, 19, false}, {13, 20, true}},
			3: {{15, 16, false}, {15, 17, true}},
		},
	}
}

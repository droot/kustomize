// Code generated by "stringer -type=loadRestrictions"; DO NOT EDIT.

package krusty

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[unknown-0]
	_ = x[rootOnly-1]
	_ = x[none-2]
}

const _loadRestrictions_name = "unknownrootOnlynone"

var _loadRestrictions_index = [...]uint8{0, 7, 15, 19}

func (i loadRestrictions) String() string {
	if i < 0 || i >= loadRestrictions(len(_loadRestrictions_index)-1) {
		return "loadRestrictions(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _loadRestrictions_name[_loadRestrictions_index[i]:_loadRestrictions_index[i+1]]
}

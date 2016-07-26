package oddsengine

import (
	"reflect"
	"testing"
)

func TestRollMapAddRoll(t *testing.T) {
	values := []struct {
		values map[int]int
		result RollMap
	}{
		{map[int]int{1: 3, 3: 2}, RollMap{{1, 3}, {3, 2}}},
	}
	for _, tt := range values {
		rMap := RollMap{}
		for hitValue, num := range tt.values {
			rMap = rMap.AddRoll(hitValue, num)
		}
		if !reflect.DeepEqual(rMap, tt.result) {
			t.Errorf("adding rolls to an empty rollmap did not create correct values\ngot: %v\nexpected: %v", rMap, tt.result)
		}
	}
}

func TestRollMapGetValue(t *testing.T) {
	values := []struct {
		rollmap    RollMap
		checkValue int
		expected   RollValue
	}{
		{RollMap{{1, 3}, {2, 2}}, 2, RollValue{2, 2}},
		{RollMap{{2, 5}, {3, 2}, {4, 11}}, 3, RollValue{3, 2}},
	}

	for _, tt := range values {
		_, actual := tt.rollmap.Get(tt.checkValue)
		if !reflect.DeepEqual(actual, &tt.expected) {
			t.Errorf("failed to get the correct rollmap value.\nexpected: %v\nactual: %v", tt.expected, actual)
		}
	}
}

func TestRollMapHasValue(t *testing.T) {
	values := []struct {
		rollmap    RollMap
		checkValue int
		exists     bool
	}{
		{RollMap{{1, 3}, {3, 2}}, 1, true},
		{RollMap{{2, 3}, {3, 2}}, 1, false},
		{RollMap{{1, 3}, {3, 2}, {4, 2}}, 2, false},
		{RollMap{{1, 3}, {3, 2}, {4, 2}}, 4, true},
	}

	for _, tt := range values {
		if tt.rollmap.HasValue(tt.checkValue) != tt.exists {
			t.Errorf("RollMap did not find a correct value\nin: %v\nfor: %v", tt.rollmap, tt.checkValue)
		}
	}
}

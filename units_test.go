package oddsengine

import (
	"reflect"
	"sort"
	"testing"
)

func TestUnitSwitching(t *testing.T) {
	values := []struct {
		game  string
		units []string
	}{
		{"1941", []string{"inf", "sub", "tan", "des", "fig", "car", "bom", "bat"}},
		{"1942", []string{"inf", "art", "sub", "tan", "des", "fig", "cru", "bom", "car", "bat", "aaa"}},
		{"1940", []string{"kam", "inf", "mec", "imec", "art", "aart", "raaa", "sub", "ssub", "tan", "des", "fig", "jfig", "tac", "cru", "hbom", "bom", "car", "bat", "aaa"}},
	}
	for _, tt := range values {
		SetGame(tt.game)
		sort.Sort(ByCost{activeUnits})
		actual := unitsToSlice(activeUnits)
		if !reflect.DeepEqual(tt.units, actual) {
			t.Errorf("Units for game %v not generated correctly\nexpected: %v\nactual: %v", tt.game, tt.units, actual)
		}
	}
}

func TestUnitsSorting(t *testing.T) {
	var actual []string
	units := Units{
		{
			Alias:  "inf",
			Name:   "Infantry",
			Cost:   3,
			Attack: 1,
			Defend: 2,
		},
		{
			Alias:      "cru",
			Name:       "Cruiser",
			Cost:       12,
			Attack:     3,
			Defend:     3,
			IsShip:     true,
			CanBombard: true,
		},
		{
			Alias:  "art",
			Name:   "Artillery",
			Cost:   4,
			Attack: 2,
			Defend: 2,
		},
		{
			Alias:       "bat",
			Name:        "Battleship",
			Cost:        20,
			Attack:      4,
			Defend:      4,
			IsShip:      true,
			CanBombard:  true,
			CapitalShip: true,
		},
		{
			Alias:       "car",
			Name:        "Aircraft Carrier",
			Cost:        14,
			Attack:      1,
			Defend:      2,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: false,
		},
	}

	byValueOrder := []string{"inf", "art", "cru", "car", "bat"}
	byAttackingOrder := []string{"inf", "car", "art", "cru", "bat"}
	byDefendingOrder := []string{"inf", "art", "car", "cru", "bat"}

	sort.Sort(ByCost{units})
	actual = unitsToSlice(units)

	if !reflect.DeepEqual(actual, byValueOrder) {
		t.Errorf("sorting units by value is not working\nexpected: %v\nactual: %v", byValueOrder, actual)
	}

	sort.Sort(ByAttackingPower{units})
	actual = unitsToSlice(units)

	if !reflect.DeepEqual(actual, byAttackingOrder) {
		t.Errorf("sorting units by attack is not working\nexpected: %v\nactual: %v", byAttackingOrder, actual)
	}

	sort.Sort(ByDefendingPower{units})
	actual = unitsToSlice(units)

	if !reflect.DeepEqual(actual, byDefendingOrder) {
		t.Errorf("sorting units by defence is not working\nexpected: %v\nactual: %v", byDefendingOrder, actual)
	}

}

func unitsToSlice(p Units) []string {
	ps := []string{}
	for _, unit := range p {
		ps = append(ps, unit.Alias)
	}

	return ps
}

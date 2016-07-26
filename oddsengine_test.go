package oddsengine

import (
	"reflect"
	"testing"
)

func TestIsResolvedFunction(t *testing.T) {
	values := []struct {
		attackers map[string]int
		defenders map[string]int
		result    bool
	}{
		{map[string]int{"tac": 2, "mec": 1, "art": 3}, map[string]int{"tan": 1}, false},
		{map[string]int{"inf": 2, "tan": 1, "fig": 2}, map[string]int{}, true},
	}
	for _, tt := range values {
		if isResolved(tt.attackers, tt.defenders) != tt.result {
			t.Errorf("\nConflict Resolution marked incorrectly.\nattackers:%v\ndefenders: %v", tt.attackers, tt.defenders)
		}
	}
}

func TestCasualtyTaking(t *testing.T) {
	values := []struct {
		units     map[string]int
		hits      int
		aftermath map[string]int
	}{
		{map[string]int{"inf": 3, "art": 2}, 4, map[string]int{"art": 1}},
		{map[string]int{"bat": 2, "car": 1, "des": 1, "sub": 2}, 5, map[string]int{"-bat": 2, "-car": 1, "des": 1}},
		{map[string]int{"bat": 2, "car": 1, "des": 1}, 2, map[string]int{"-bat": 1, "-car": 1, "bat": 1, "des": 1}},
		{map[string]int{"tan": 4, "fig": 1, "inf": 1, "art": 3}, 7, map[string]int{"fig": 1, "tan": 1}},
		{map[string]int{"tan": 2, "fig": 1, "inf": 1, "art": 3}, 7, map[string]int{}},
	}
	for _, tt := range values {
		takeCasualties(tt.units, tt.hits, baseOol)

		if !reflect.DeepEqual(tt.aftermath, tt.units) {
			t.Errorf("casualties did not take properly\nexpected: %v\nactual:%v", tt.aftermath, tt.units)
		}
	}
}

func TestHasUndamagedCapitalShip(t *testing.T) {
	values := []struct {
		units      map[string]int
		hasCapital bool
	}{
		{map[string]int{"car": 2, "sub": 1, "des": 3}, true},
		{map[string]int{"cru": 2, "sub": 1, "des": 3}, false},
		{map[string]int{"bat": 1, "cru": 2, "sub": 1, "des": 3}, true},
		{map[string]int{"car": 3, "bat": 1, "cru": 2, "sub": 1, "des": 3}, true},
		{map[string]int{"-bat": 1, "cru": 2, "sub": 1, "des": 3}, false},
	}
	for _, tt := range values {
		if hasUndamagedCapitalShips(tt.units) != tt.hasCapital {
			t.Errorf("Unit's Capital Ships marked incorrectly\n%v\n", tt.units)
		}
	}
}

func TestCapitalShipDamage(t *testing.T) {
	values := []struct {
		units      map[string]int
		hits       int
		aftermath  map[string]int
		numDamaged int
	}{
		{map[string]int{"car": 2, "sub": 1, "des": 3}, 2, map[string]int{"-car": 2, "sub": 1, "des": 3}, 2},
		{map[string]int{"cru": 2, "sub": 1, "des": 3}, 3, map[string]int{"cru": 2, "sub": 1, "des": 3}, 0},
		{map[string]int{"bat": 1, "car": 2, "sub": 1, "des": 3}, 1, map[string]int{"bat": 1, "-car": 1, "car": 1, "sub": 1, "des": 3}, 1},
		{map[string]int{"car": 1, "bat": 2, "cru": 2, "sub": 1, "des": 3}, 2, map[string]int{"-car": 1, "-bat": 1, "bat": 1, "cru": 2, "sub": 1, "des": 3}, 2},
	}
	for _, tt := range values {
		numDamaged := damageCapitalShips(tt.units, tt.hits)
		if !reflect.DeepEqual(tt.aftermath, tt.units) {
			t.Errorf("we did not damage the capital ships correctly\nexpected: %v\nactual: %v\n", tt.aftermath, tt.units)
		}
		if numDamaged != tt.numDamaged {
			t.Errorf("we did not report the number of capital damage taken correctly\n%v\n", tt.units)
		}
	}
}

func TestSupriseAttack(t *testing.T) {
	values := []struct {
		attackers   map[string]int
		defenders   map[string]int
		attackerCan bool
		defenderCan bool
	}{
		{map[string]int{"sub": 1, "cru": 2}, map[string]int{"sub": 2, "bat": 2}, true, true},
		{map[string]int{"sub": 1, "cru": 2}, map[string]int{"sub": 2, "des": 1}, false, true},
		{map[string]int{"des": 1, "cru": 2}, map[string]int{"sub": 2, "des": 1}, false, false},
		{map[string]int{"des": 1, "cru": 2}, map[string]int{"car": 2, "des": 1}, false, false},
	}

	for _, tt := range values {
		if attackerCanSupriseAttack(tt.attackers, tt.defenders) != tt.attackerCan {
			t.Errorf("the attacker suprise attack ability was not calculated correctly")
		}
		if defenderCanSupriseAttack(tt.attackers, tt.defenders) != tt.defenderCan {
			t.Errorf("the attacker suprise attack ability was not calculated correctly")
		}
	}
}

func TestCanBombard(t *testing.T) {
	values := []struct {
		units      map[string]int
		can        bool
		bombardMap RollMap
	}{
		{map[string]int{"sub": 1, "car": 3}, false, RollMap{}},
		{map[string]int{"sub": 1, "bat": 1, "inf": 1}, true, RollMap{{4, 1}}},
		{map[string]int{"sub": 1, "des": 3}, false, RollMap{}},
		{map[string]int{"cru": 1, "des": 3, "tan": 2}, true, RollMap{{3, 1}}},
		{map[string]int{"car": 2, "cru": 1, "des": 3}, false, RollMap{}},
		{map[string]int{"car": 2, "bat": 1, "cru": 1, "des": 3, "inf": 1, "art": 2}, true, RollMap{{3, 1}, {4, 1}}},
	}

	for _, tt := range values {
		if canBombard(tt.units) != tt.can {
			t.Errorf("Units bombard not calculated correctly.\n%v", tt.units)
		}

		if tt.can {
			if !reflect.DeepEqual(getBombardRollMap(tt.units), tt.bombardMap) {
				t.Errorf("Bombard Rollmap not created correctly\nexpected: %v\nactual: %v\n", tt.bombardMap, getBombardRollMap(tt.units))
			}
		}
	}

}

func TestReservationOfHighestValueLandUnit(t *testing.T) {
	values := []struct {
		units    map[string]int
		expected map[string]int
	}{
		{map[string]int{"inf": 1, "art": 1, "tan": 1, "fig": 3}, map[string]int{"inf": 1, "art": 1, "+tan": 1, "fig": 3}},
		{map[string]int{"inf": 1, "art": 1, "fig": 3}, map[string]int{"inf": 1, "+art": 1, "fig": 3}},
	}
	for _, tt := range values {
		reserveHighestValueLandUnit(tt.units)
		if !reflect.DeepEqual(tt.expected, tt.units) {
			t.Errorf("reserving the highest value land unit doesn't work right\nexpected: %v\nactual: %v\n", tt.expected, tt.units)
		}
	}

}

func TestMecAndInfPlusOneFunc(t *testing.T) {
	values := []struct {
		units         map[string]int
		numInfBoosted int
		numMecBoosted int
	}{
		{map[string]int{"inf": 2, "mec": 1, "art": 4}, 2, 1},
		{map[string]int{"inf": 2, "mec": 1}, 0, 0},
		{map[string]int{"inf": 2, "mec": 1, "art": 2, "fig": 4, "tan": 1}, 2, 0},
	}

	inf := activePieces.Find("inf")
	mec := activePieces.Find("mec")

	for _, tt := range values {
		if inf.PlusOneShots(tt.units) != tt.numInfBoosted {
			t.Errorf("did not return correct infantry \"plus one\" shots\n%v", tt.units)
		}
		if mec.PlusOneShots(tt.units) != tt.numMecBoosted {
			t.Errorf("did not return correct mec \"plus one\" shots\n%v", tt.units)
		}
	}

}

func TestTacPlusOneFunc(t *testing.T) {
	values := []struct {
		units         map[string]int
		numTacBoosted int
	}{
		{map[string]int{"tan": 2, "fig": 1, "tac": 4}, 3},
		{map[string]int{"tac": 2, "mec": 1, "art": 3}, 0},
		{map[string]int{"tac": 2, "mec": 1, "art": 2, "fig": 4, "tan": 1}, 2},
		{map[string]int{"tan": 2, "tac": 1}, 1},
	}

	tac := activePieces.Find("tac")

	for _, tt := range values {
		if tac.PlusOneShots(tt.units) != tt.numTacBoosted {
			t.Errorf("did not return correct Tac \"plus one\" shots\n%v", tt.units)
		}
	}
}

func TestRollMapper(t *testing.T) {
	values := []struct {
		units    map[string]int
		mode     string
		expected RollMap
	}{
		{map[string]int{"tac": 2, "mec": 1, "art": 3}, "attack", RollMap{{2, 4}, {3, 2}}},
		{map[string]int{"tac": 3, "fig": 1, "tan": 1}, "attack", RollMap{{3, 3}, {4, 2}}},
		{map[string]int{"sub": 3, "car": 1, "bat": 1, "des": 2}, "defend", RollMap{{1, 3}, {2, 3}, {4, 1}}},
		{map[string]int{"sub": 3, "car": 1, "bat": 1, "des": 2}, "attack", RollMap{{0, 1}, {2, 5}, {4, 1}}},
		{map[string]int{"inf": 8, "art": 10, "mec": 5, "fig": 2}, "attack", RollMap{{1, 3}, {2, 20}, {3, 2}}},
		{map[string]int{"kam": 1, "des": 4, "car": 1}, "defend", RollMap{{2, 6}}},
	}
	for _, tt := range values {

		rmap := createRollMap(tt.units, tt.mode)

		if !reflect.DeepEqual(tt.expected, rmap) {
			t.Errorf("roll map did not generate correctly\nexpected:%v\nactual:%v", tt.expected, rmap)
		}
	}
}

func TestSetOol(t *testing.T) {
	testOol := []string{"inf", "tan", "art"}

	SetBaseOol(testOol)

	if !reflect.DeepEqual(testOol, baseOol) {
		t.Errorf("setting the baseOol failed")
	}
}

func TestPieceValidity(t *testing.T) {
	values := []struct {
		pieces map[string]int
		game   string
		valid  bool
	}{
		{map[string]int{"inf": 1, "tac": 2}, "1941", false},
		{map[string]int{"inf": 1, "mec": 2, "kam": 1}, "1942", false},
		{map[string]int{"unk": 1, "mec": 2, "kam": 1}, "1940", false},
		{map[string]int{"inf": 1, "mec": 2, "kam": 1, "bat": 2}, "1940", true},
		{map[string]int{"+inf": 1, "mec": 2, "kam": 1, "bat": 2}, "1940", true},
		{map[string]int{"inf": 1, "mec": 2, "kam": 1, "-bat": 2}, "1940", true},
		{map[string]int{"+unk": 1, "mec": 2, "kam": 1, "-bat": 2}, "1941", false},
	}

	for _, tt := range values {
		SetGame(tt.game)
		err := checkPieceValidity(tt.pieces)
		if (err == nil) != tt.valid {
			t.Errorf("Piece validity was not determined correctly for game %v\npieces: %v\nmessage: %v", tt.game, tt.pieces, err)
		}
	}
}

func TestHasLimitedAircraft(t *testing.T) {
	values := []struct {
		attackers map[string]int
		defenders map[string]int
		result    bool
	}{
		{map[string]int{"fig": 1}, map[string]int{"sub": 2}, true},
		{map[string]int{"fig": 1, "des": 1}, map[string]int{"sub": 2}, false},
		{map[string]int{"bat": 1, "tac": 2, "sub": 2}, map[string]int{"bat": 1, "tac": 2, "sub": 1}, true},
		{map[string]int{"des": 1, "bom": 1, "sub": 2}, map[string]int{"bat": 1, "fig": 1, "des": 1}, false},
	}
	for _, tt := range values {
		if hasLimitedAircraft(tt.attackers, tt.defenders) != tt.result {
			t.Errorf("The need to limit aircraft not calculated correctly\nattackers: %v\ndefenders: %v", tt.attackers, tt.defenders)
		}
	}
}

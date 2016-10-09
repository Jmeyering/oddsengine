package oddsengine

import (
	"math/rand"
	"reflect"
	"testing"
)

// TestConflictResolution test conflict with a given set of attackers, and
// defenders. In order to ensure the battle is a known quantity, we seed each
// battle with a different random seed. You can view the result of seeding
// random with different random Seeds here
// https://play.golang.org/p/1L-kbymrnK
// look for specific results here
// https://play.golang.org/p/EmS8-b7cV-
//
// The order of the rolls is rather complicated.
//
// 1.  Roll Kamakazi
// 2.  Roll AAA
// 3.  Roll Bombard
// 4.  Roll Attacker Sub Suprise Attack
// 5.  Roll Defender Sub Suprise Attack
// 6.  Roll Attacker Aircraft
// 7.  Roll Attacker Subs
// 8.  Roll Remaining Attacker Units
// 9.  Roll Defender Aircraft
// 10. Roll Defender Subs
// 11. Roll Remaining Defender Units
func TestConflictResolution(t *testing.T) {
	values := []struct {
		attackers         map[string]int
		defenders         map[string]int
		game              string
		randSeed          int64
		outcome           ConflictProfile
		mustTakeTerritory bool
	}{
		{
			map[string]int{"tac": 2, "mec": 1, "art": 3},
			map[string]int{"inf": 4, "tan": 2, "fig": 3},
			"1940",
			1,
			ConflictProfile{
				Rounds:                 2,
				DefenderHits:           []int{5, 3},
				AttackerHits:           []int{2, 1},
				AttackerIpcLoss:        38,
				DefenderIpcLoss:        9,
				AAAHits:                0,
				KamikazeHits:           0,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"inf": 1, "tan": 2, "fig": 3}),
				Outcome:                -1,
			},
			false,
		},
		// Test with boosted MEC and a bombard
		{
			map[string]int{"mec": 1, "art": 3, "tan": 1, "tac": 2, "bat": 2},
			map[string]int{"inf": 3, "mec": 3, "tan": 1},
			"1940",
			1,
			ConflictProfile{
				Rounds:                 3,
				AttackerHits:           []int{4, 0, 1},
				DefenderHits:           []int{4, 1, 2},
				AttackerIpcLoss:        44,
				DefenderIpcLoss:        17,
				AAAHits:                0,
				KamikazeHits:           0,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"mec": 1, "tan": 1}),
				Outcome:                -1,
			},
			false,
		},
		// Test AAA With Bombard
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2, "bom": 1},
			map[string]int{"aaa": 1, "inf": 2, "tan": 2},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{1},
				AttackerHits:           []int{5},
				AttackerIpcLoss:        23,
				DefenderIpcLoss:        23,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"inf": 3, "fig": 1, "tac": 2, "bom": 1}),
				AAAHits:                2,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		{
			map[string]int{"sub": 2, "bat": 1, "cru": 1, "fig": 1},
			map[string]int{"sub": 1, "cru": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{3},
				AttackerHits:           []int{4},
				AttackerIpcLoss:        12,
				DefenderIpcLoss:        42,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"-bat": 1, "cru": 1, "fig": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Testing a carrier capital ship damage
		{
			map[string]int{"sub": 2, "bat": 1, "des": 1, "fig": 1},
			map[string]int{"sub": 1, "cru": 3, "car": 1},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 3,
				AttackerHits:           []int{3, 2, 2},
				DefenderHits:           []int{3, 1, 1},
				AttackerIpcLoss:        30,
				DefenderIpcLoss:        58,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"-bat": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Test defender suprise attacks
		{
			map[string]int{"cru": 4, "fig": 1},
			map[string]int{"sub": 2, "cru": 1, "bat": 1},
			"1940",
			5,
			ConflictProfile{
				Rounds:                 3,
				DefenderHits:           []int{2, 1, 1},
				AttackerHits:           []int{1, 2, 2},
				AttackerIpcLoss:        46,
				DefenderIpcLoss:        44,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"cru": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		{
			map[string]int{"des": 1},
			map[string]int{"des": 1, "kam": 2},
			"1940",
			5,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{1},
				AttackerHits:           []int{0},
				AttackerIpcLoss:        8,
				DefenderIpcLoss:        0,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"des": 1}),
				AAAHits:                0,
				KamikazeHits:           1,
				Outcome:                -1,
			},
			false,
		},
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2},
			map[string]int{"aaa": 2, "inf": 2, "tan": 2},
			"1940",
			3,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{2},
				AttackerHits:           []int{4},
				AttackerIpcLoss:        26,
				DefenderIpcLoss:        28,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"inf": 2, "fig": 1, "tac": 2}),
				AAAHits:                2,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		{
			map[string]int{"fig": 3},
			map[string]int{"tan": 3},
			"1940",
			19,
			ConflictProfile{
				Rounds:          3,
				DefenderHits:    []int{1, 1, 1},
				AttackerHits:    []int{1, 1, 1},
				AttackerIpcLoss: 30,
				DefenderIpcLoss: 18,
				AAAHits:         0,
				KamikazeHits:    0,
				Outcome:         0,
			},
			false,
		},
		// Testing with must take territory.
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2},
			map[string]int{"aaa": 2, "inf": 4, "tan": 2},
			"1940",
			3,
			ConflictProfile{
				Rounds:                 5,
				AttackerHits:           []int{4, 1, 0, 0, 0},
				DefenderHits:           []int{4, 1, 0, 1, 1},
				AttackerIpcLoss:        64,
				DefenderIpcLoss:        18,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"aaa": 2, "tan": 1}),
				AAAHits:                2,
				KamikazeHits:           0,
				Outcome:                -1,
			},
			true,
		},
		{
			map[string]int{"inf": 4, "art": 3},
			map[string]int{"inf": 1, "tan": 3},
			"1940",
			1,
			ConflictProfile{
				Rounds:                 2,
				DefenderHits:           []int{2, 0},
				AttackerHits:           []int{3, 2},
				AttackerIpcLoss:        6,
				DefenderIpcLoss:        21,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"inf": 2, "art": 3}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Test that a reserved unit is taken last when attacking
		{
			map[string]int{"inf": 4, "fig": 2, "+mec": 1},
			map[string]int{"inf": 3, "tan": 1},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 3,
				AttackerHits:           []int{2, 0, 2},
				DefenderHits:           []int{4, 0, 2},
				AttackerIpcLoss:        32,
				DefenderIpcLoss:        15,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"+mec": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Test of combination of reserved attackers and defenders
		{
			map[string]int{"+inf": 1, "inf": 2, "art": 3, "fig": 2},
			map[string]int{"+inf": 1, "mec": 5, "tan": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 3,
				AttackerHits:           []int{3, 3, 0},
				DefenderHits:           []int{4, 3, 1},
				AttackerIpcLoss:        41,
				DefenderIpcLoss:        26,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"+inf": 1, "tan": 2}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                -1,
			},
			false,
		},
		// Test that a reserved unit is taken last when defending
		{
			map[string]int{"inf": 2, "art": 3, "tan": 2},
			map[string]int{"+inf": 1, "mec": 5, "tan": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 2,
				DefenderHits:           []int{4, 3},
				AttackerHits:           []int{3, 3},
				AttackerIpcLoss:        30,
				DefenderIpcLoss:        26,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"+inf": 1, "tan": 2}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                -1,
			},
			false,
		},
		{
			map[string]int{"+sub": 1, "sub": 3, "des": 6, "cru": 1, "car": 1, "fig": 2},
			map[string]int{"kam": 1, "des": 2, "bat": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 2,
				DefenderHits:           []int{3, 1},
				AttackerHits:           []int{5, 5},
				AttackerIpcLoss:        18,
				DefenderIpcLoss:        76,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"-car": 1, "fig": 2, "+sub": 1, "des": 6, "cru": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		{
			map[string]int{"sub": 4, "des": 6, "cru": 1, "car": 1, "fig": 2},
			map[string]int{"kam": 1, "des": 2, "bat": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                 2,
				DefenderHits:           []int{3, 1},
				AttackerHits:           []int{5, 5},
				AttackerIpcLoss:        18,
				DefenderIpcLoss:        76,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"-car": 1, "fig": 2, "sub": 1, "des": 6, "cru": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Test that a draw is recorded for attacks involving unhittable
		// submarines and attacking fighters
		{
			map[string]int{"fig": 4},
			map[string]int{"sub": 10},
			"1940",
			1,
			ConflictProfile{
				Rounds:                 0,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"fig": 4}),
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"sub": 10}),
				Outcome:                0,
			},
			false,
		},
		// Test that in 1942 Carriers are not represented as capital ships.
		{
			map[string]int{"car": 1},
			map[string]int{"sub": 1},
			"1942",
			5,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{1},
				AttackerHits:           []int{0},
				AttackerIpcLoss:        14,
				DefenderIpcLoss:        0,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"sub": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                -1,
			},
			false,
		},
		// Test that AAA are counted as defending units, when there are only
		// planes attacking a region. Generally an attacker on AAA is an auto
		// kill, this is an exception to that rule.
		{
			map[string]int{"fig": 1},
			map[string]int{"aaa": 1},
			"1940",
			5,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{0},
				AttackerHits:           []int{0},
				AttackerIpcLoss:        10,
				DefenderIpcLoss:        0,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"aaa": 1}),
				AAAHits:                1,
				KamikazeHits:           0,
				Outcome:                -1,
			},
			false,
		},
		{
			map[string]int{"fig": 2},
			map[string]int{"aaa": 1},
			"1940",
			51,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{0},
				AttackerHits:           []int{0},
				AttackerIpcLoss:        10,
				DefenderIpcLoss:        5,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"fig": 1}),
				AAAHits:                1,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		{
			map[string]int{"fig": 3},
			map[string]int{"aaa": 1},
			"1940",
			98,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{0},
				AttackerHits:           []int{1},
				AttackerIpcLoss:        10,
				DefenderIpcLoss:        5,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"fig": 2}),
				AAAHits:                1,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		{
			map[string]int{"fig": 3},
			map[string]int{"raaa": 1},
			"1940",
			98,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{0},
				AttackerHits:           []int{0},
				AttackerIpcLoss:        30,
				DefenderIpcLoss:        0,
				DefenderUnitsRemaining: formationToSortedSlice(map[string]int{"raaa": 1}),
				AAAHits:                3,
				KamikazeHits:           0,
				Outcome:                -1,
			},
			false,
		},
		{
			map[string]int{"inf": 1, "fig": 3},
			map[string]int{"raaa": 1},
			"1940",
			98,
			ConflictProfile{
				Rounds:                 1,
				DefenderHits:           []int{0},
				AttackerHits:           []int{0},
				AttackerIpcLoss:        30,
				DefenderIpcLoss:        5,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"inf": 1}),
				AAAHits:                3,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Multi roll unit test
		{
			map[string]int{"hbom": 2, "jfig": 3},
			map[string]int{"raaa": 1, "aart": 2, "imec": 2, "tan": 1},
			"1940",
			10,
			ConflictProfile{
				Rounds:                 3,
				AttackerHits:           []int{3, 1, 2},
				DefenderHits:           []int{1, 1, 0},
				AttackerIpcLoss:        30,
				DefenderIpcLoss:        27,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"hbom": 2}),
				AAAHits:                1,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// Ensure that multi hit aircraft still are limited against unhittable
		// subs. The planes should not roll any dice for this test.
		{
			map[string]int{"hbom": 1, "sub": 1},
			map[string]int{"sub": 1},
			"1940",
			1,
			ConflictProfile{
				Rounds:                 3,
				AttackerHits:           []int{0, 0, 1},
				DefenderHits:           []int{0, 0, 1},
				AttackerIpcLoss:        6,
				DefenderIpcLoss:        6,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"hbom": 1}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
		// This check makes sure that reserved defending units actually get reserved
		{
			map[string]int{"tac": 10, "cru": 1},
			map[string]int{"car": 1, "+des": 1},
			"1940",
			1,
			ConflictProfile{
				Rounds:                 1,
				AttackerHits:           []int{6},
				DefenderHits:           []int{2},
				AttackerIpcLoss:        22,
				DefenderIpcLoss:        24,
				AttackerUnitsRemaining: formationToSortedSlice(map[string]int{"cru": 1, "tac": 8}),
				AAAHits:                0,
				KamikazeHits:           0,
				Outcome:                1,
			},
			false,
		},
	}
	for _, tt := range values {
		SetGame(tt.game)
		SetMustTakeTerritory(tt.mustTakeTerritory)
		if mustTakeTerritory {
			reserveHighestValueLandUnit(tt.attackers)
		}

		ool := customizeOol(tt.attackers, tt.defenders)
		rand.Seed(tt.randSeed)
		p := resolveConflict(tt.attackers, tt.defenders, ool)
		if !reflect.DeepEqual(p, &tt.outcome) {
			t.Errorf("Conflict Profile Doesn't Match\nexpected: %+v\nactual: %+v", tt.outcome, *p)
		}
	}

	// Reset the game back to 1940
	SetGame("1940")
}

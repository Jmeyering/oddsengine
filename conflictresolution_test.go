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
// 6.  Roll Attacker Aircraft (If there are defending sub's which the aircraft cannot hit.)
// 7.  Roll Attacker Subs
// 8.  Roll Remaining Attacker Pieces
// 9.  Roll Defender Aircraft (If there are attacking sub's which the aircraft cannot hit.)
// 10. Roll Defender Subs
// 11. Roll Remaining Defender Pieces
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
				Rounds:                  2,
				DefenderHits:            []int{5, 3},
				AttackerHits:            []int{2, 1},
				AttackerIpcLoss:         38,
				DefenderIpcLoss:         9,
				AAAHits:                 0,
				KamikazeHits:            0,
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"inf": 1, "tan": 2, "fig": 3}),
				Outcome:                 -1,
			},
			false,
		},
		{
			map[string]int{"mec": 1, "art": 3, "tan": 1, "tac": 2, "bat": 2},
			map[string]int{"inf": 3, "mec": 3, "tan": 1},
			"1940",
			1,
			ConflictProfile{
				Rounds:                  3,
				DefenderHits:            []int{4, 0, 1},
				AttackerHits:            []int{5, 0, 2},
				AttackerIpcLoss:         22,
				DefenderIpcLoss:         27,
				AAAHits:                 0,
				KamikazeHits:            0,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"tac": 2}),
				Outcome:                 1,
			},
			false,
		},
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2, "bom": 1},
			map[string]int{"aaa": 1, "inf": 2, "tan": 2},
			"1940",
			2,
			ConflictProfile{
				Rounds:                  1,
				DefenderHits:            []int{1},
				AttackerHits:            []int{4},
				AttackerIpcLoss:         23,
				DefenderIpcLoss:         23,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"inf": 3, "fig": 1, "tac": 2, "bom": 1}),
				AAAHits:                 2,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			false,
		},
		{
			map[string]int{"sub": 2, "bat": 1, "cru": 1, "fig": 1},
			map[string]int{"sub": 1, "cru": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                  1,
				DefenderHits:            []int{3},
				AttackerHits:            []int{4},
				AttackerIpcLoss:         12,
				DefenderIpcLoss:         42,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"-bat": 1, "cru": 1, "fig": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			false,
		},
		{
			map[string]int{"sub": 2, "bat": 1, "des": 1, "fig": 1},
			map[string]int{"sub": 1, "cru": 3, "car": 1},
			"1940",
			2,
			ConflictProfile{
				Rounds:                  3,
				DefenderHits:            []int{3, 1, 1},
				AttackerHits:            []int{4, 1, 1},
				AttackerIpcLoss:         30,
				DefenderIpcLoss:         58,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"-bat": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			false,
		},
		{
			map[string]int{"cru": 4, "fig": 1},
			map[string]int{"sub": 2, "cru": 1, "bat": 1},
			"1940",
			5,
			ConflictProfile{
				Rounds:                  3,
				DefenderHits:            []int{2, 1, 1},
				AttackerHits:            []int{1, 2, 2},
				AttackerIpcLoss:         46,
				DefenderIpcLoss:         44,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"cru": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			false,
		},
		{
			map[string]int{"des": 1},
			map[string]int{"des": 1, "kam": 2},
			"1940",
			5,
			ConflictProfile{
				Rounds:                  1,
				DefenderHits:            []int{1},
				AttackerHits:            []int{0},
				AttackerIpcLoss:         8,
				DefenderIpcLoss:         0,
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"des": 1}),
				AAAHits:                 0,
				KamikazeHits:            1,
				Outcome:                 -1,
			},
			false,
		},
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2},
			map[string]int{"aaa": 2, "inf": 2, "tan": 2},
			"1940",
			3,
			ConflictProfile{
				Rounds:                  1,
				DefenderHits:            []int{2},
				AttackerHits:            []int{4},
				AttackerIpcLoss:         26,
				DefenderIpcLoss:         28,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"inf": 2, "fig": 1, "tac": 2}),
				AAAHits:                 2,
				KamikazeHits:            0,
				Outcome:                 1,
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
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2},
			map[string]int{"aaa": 2, "inf": 4, "tan": 2},
			"1940",
			3,
			ConflictProfile{
				Rounds:                  4,
				DefenderHits:            []int{4, 1, 0, 0},
				AttackerHits:            []int{4, 0, 1, 1},
				AttackerIpcLoss:         50,
				DefenderIpcLoss:         34,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"+inf": 1, "tac": 1}),
				AAAHits:                 2,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			true,
		},
		{
			map[string]int{"inf": 4, "art": 3},
			map[string]int{"inf": 1, "tan": 3},
			"1940",
			1,
			ConflictProfile{
				Rounds:                  2,
				DefenderHits:            []int{2, 0},
				AttackerHits:            []int{3, 2},
				AttackerIpcLoss:         6,
				DefenderIpcLoss:         21,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"inf": 2, "art": 3}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			false,
		},
		// Test that a reserved unit is taken last when attacking
		{
			map[string]int{"inf": 4, "fig": 2, "+mec": 1},
			map[string]int{"inf": 2, "tan": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                  3,
				DefenderHits:            []int{4, 1, 1},
				AttackerHits:            []int{3, 1, 2},
				AttackerIpcLoss:         32,
				DefenderIpcLoss:         24,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"+mec": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 1,
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
				Rounds:                  5,
				DefenderHits:            []int{3, 2, 1, 1, 1},
				AttackerHits:            []int{4, 2, 1, 0, 0},
				AttackerIpcLoss:         41,
				DefenderIpcLoss:         32,
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"+inf": 1, "tan": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 -1,
			},
			false,
		},
		// Test that a reserved unit is taken last when defending
		{
			map[string]int{"inf": 2, "art": 3, "fig": 2},
			map[string]int{"+inf": 1, "mec": 5, "tan": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                  2,
				DefenderHits:            []int{4, 3},
				AttackerHits:            []int{3, 3},
				AttackerIpcLoss:         38,
				DefenderIpcLoss:         26,
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"+inf": 1, "tan": 2}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 -1,
			},
			false,
		},
		{
			map[string]int{"sub": 4, "des": 6, "cru": 1, "car": 1, "fig": 2},
			map[string]int{"kam": 1, "des": 2, "bat": 3},
			"1940",
			2,
			ConflictProfile{
				Rounds:                  2,
				DefenderHits:            []int{3, 1},
				AttackerHits:            []int{5, 5},
				AttackerIpcLoss:         18,
				DefenderIpcLoss:         76,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"-car": 1, "fig": 2, "sub": 1, "des": 6, "cru": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 1,
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
				Rounds:                  0,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"fig": 4}),
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"sub": 10}),
				Outcome:                 0,
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
				Rounds:                  1,
				DefenderHits:            []int{1},
				AttackerHits:            []int{0},
				AttackerIpcLoss:         14,
				DefenderIpcLoss:         0,
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"sub": 1}),
				AAAHits:                 0,
				KamikazeHits:            0,
				Outcome:                 -1,
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
				Rounds:                  1,
				DefenderHits:            []int{0},
				AttackerHits:            []int{0},
				AttackerIpcLoss:         10,
				DefenderIpcLoss:         0,
				DefenderPiecesRemaining: formationToSortedSlice(map[string]int{"aaa": 1}),
				AAAHits:                 1,
				KamikazeHits:            0,
				Outcome:                 -1,
			},
			false,
		},
		{
			map[string]int{"fig": 2},
			map[string]int{"aaa": 1},
			"1940",
			51,
			ConflictProfile{
				Rounds:                  1,
				DefenderHits:            []int{0},
				AttackerHits:            []int{0},
				AttackerIpcLoss:         10,
				DefenderIpcLoss:         5,
				AttackerPiecesRemaining: formationToSortedSlice(map[string]int{"fig": 1}),
				AAAHits:                 1,
				KamikazeHits:            0,
				Outcome:                 1,
			},
			false,
		},
	}
	for _, tt := range values {
		SetGame(tt.game)
		SetMustTakeTerritory(tt.mustTakeTerritory)

		ool := customizeOol(tt.attackers, tt.defenders)
		if mustTakeTerritory {
			reserveHighestValueLandUnit(tt.attackers)
		}
		rand.Seed(tt.randSeed)
		p := resolveConflict(tt.attackers, tt.defenders, ool)
		if !reflect.DeepEqual(p, &tt.outcome) {
			t.Errorf("Conflict Profile Doesn't Match\nexpected: %+v\nactual: %+v", tt.outcome, *p)
		}
	}

	// Reset the game back to 1940
	SetGame("1940")
}

func TestGetSummary(t *testing.T) {
	values := []struct {
		attackers         map[string]int
		defenders         map[string]int
		game              string
		randSeed          int64
		iterations        int
		outcome           Summary
		mustTakeTerritory bool
	}{
		{
			map[string]int{"inf": 4, "fig": 3, "tac": 2, "bom": 1},
			map[string]int{"aaa": 1, "inf": 2, "tan": 2},
			"1940",
			2,
			1,
			Summary{
				AverageRounds:           float64(1),
				AttackerWinPercentage:   float64(100),
				DefenderWinPercentage:   float64(0),
				DrawPercentage:          float64(0),
				AAAHitsAverage:          float64(2),
				KamikazeHitsAverage:     float64(0),
				AttackerAvgIpcLoss:      float64(23),
				DefenderAvgIpcLoss:      float64(23),
				AttackerPiecesRemaining: map[string]int{"bom:1,fig:1,inf:3,tac:2": 1},
				DefenderPiecesRemaining: map[string]int{},
			},
			false,
		},
		{
			map[string]int{"inf": 2, "fig": 5, "art": 3},
			map[string]int{"inf": 6, "tan": 2, "tac": 1},
			"1940",
			2,
			5,
			Summary{
				AverageRounds:           float64(3.6),
				AttackerWinPercentage:   float64(80),
				DefenderWinPercentage:   float64(20),
				DrawPercentage:          float64(0),
				AAAHitsAverage:          float64(0),
				KamikazeHitsAverage:     float64(0),
				AttackerAvgIpcLoss:      float64(38.8),
				DefenderAvgIpcLoss:      float64(38.8),
				AttackerPiecesRemaining: map[string]int{"+art:1,fig:4": 2, "+art:1,fig:3": 1, "+art:1,fig:2": 1},
				DefenderPiecesRemaining: map[string]int{"tac:1": 1},
			},
			true,
		},
		// Test a conplicated battle involving submarines under both suprise and
		// standard attack conditions
		{
			map[string]int{"sub": 4, "des": 6, "cru": 1, "car": 1, "fig": 2},
			map[string]int{"kam": 1, "des": 2, "bat": 3},
			"1940",
			2,
			6,
			Summary{
				AverageRounds:           float64(2.5),
				AttackerWinPercentage:   float64(100),
				DefenderWinPercentage:   float64(0),
				DrawPercentage:          float64(0),
				AAAHitsAverage:          float64(0),
				KamikazeHitsAverage:     float64(0.33),
				AttackerAvgIpcLoss:      float64(29),
				DefenderAvgIpcLoss:      float64(76),
				AttackerPiecesRemaining: map[string]int{"-car:1,cru:1,des:6,fig:2,sub:1": 3, "-car:1,cru:1,des:2,fig:2": 1, "-car:1,cru:1,des:6,fig:2": 1, "-car:1,cru:1,des:4,fig:2": 1},
				DefenderPiecesRemaining: map[string]int{},
			},
			false,
		},
		{
			map[string]int{"inf": 7, "tan": 1, "fig": 7, "tac": 7, "bom": 2},
			map[string]int{"inf": 15, "tan": 4, "fig": 5},
			"1940",
			3,
			10,
			Summary{
				AverageRounds:           float64(3.2),
				AttackerWinPercentage:   float64(60),
				DefenderWinPercentage:   float64(30),
				DrawPercentage:          float64(10),
				AAAHitsAverage:          float64(0),
				KamikazeHitsAverage:     float64(0),
				AttackerAvgIpcLoss:      float64(159.1),
				DefenderAvgIpcLoss:      float64(108),
				AttackerPiecesRemaining: map[string]int{"+tan:1,bom:2,tac:1": 2, "+tan:1,bom:2,tac:2": 1, "+tan:1,bom:2,tac:3": 1, "+tan:1,bom:2,tac:6": 2},
				DefenderPiecesRemaining: map[string]int{"fig:4": 2, "fig:3": 1},
			},
			true,
		},
	}

	for _, tt := range values {
		SetGame(tt.game)
		SetIterations(tt.iterations)
		SetMustTakeTerritory(tt.mustTakeTerritory)
		rand.Seed(tt.randSeed)
		s, _ := GetSummary(tt.attackers, tt.defenders)
		if !reflect.DeepEqual(s, &tt.outcome) {
			t.Errorf("Summary is incorrect.\nexpected: %+v\nactual: %+v", tt.outcome, s)
		}
	}

	// Reset the game back to 1940
	SetGame("1940")
}

package oddsengine

import (
	"reflect"
	"testing"
)

func TestAddingFirstRoundResult(t *testing.T) {
	values := []struct {
		initialSet FirstRoundResultCollection
		resultSet  FirstRoundResultCollection
		addition   FirstRoundResult
	}{
		{
			FirstRoundResultCollection{
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 6,
					Frequency:    1,
					AttackerWin:  0,
					DefenderWin:  1,
					Draw:         0,
				},
			},
			FirstRoundResultCollection{
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 6,
					Frequency:    2,
					AttackerWin:  1,
					DefenderWin:  1,
					Draw:         0,
				},
			},
			FirstRoundResult{
				AttackerHits: 6,
				DefenderHits: 6,
				Frequency:    1,
				AttackerWin:  1,
				DefenderWin:  0,
				Draw:         0,
			},
		},
		{
			FirstRoundResultCollection{
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 5,
					Frequency:    1,
					AttackerWin:  0,
					DefenderWin:  1,
					Draw:         0,
				},
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 6,
					Frequency:    1,
					AttackerWin:  1,
					DefenderWin:  0,
					Draw:         0,
				},
			},
			FirstRoundResultCollection{
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 5,
					Frequency:    2,
					AttackerWin:  1,
					DefenderWin:  1,
					Draw:         0,
				},
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 6,
					Frequency:    1,
					AttackerWin:  1,
					DefenderWin:  0,
					Draw:         0,
				},
			},
			FirstRoundResult{
				AttackerHits: 6,
				DefenderHits: 5,
				Frequency:    1,
				AttackerWin:  1,
				DefenderWin:  0,
				Draw:         0,
			},
		},
		{
			FirstRoundResultCollection{
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 5,
					Frequency:    1,
					AttackerWin:  0,
					DefenderWin:  1,
					Draw:         0,
				},
			},
			FirstRoundResultCollection{
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 5,
					Frequency:    1,
					AttackerWin:  0,
					DefenderWin:  1,
					Draw:         0,
				},
				FirstRoundResult{
					AttackerHits: 6,
					DefenderHits: 6,
					Frequency:    1,
					AttackerWin:  1,
					DefenderWin:  0,
					Draw:         0,
				},
			},
			FirstRoundResult{
				AttackerHits: 6,
				DefenderHits: 6,
				Frequency:    1,
				AttackerWin:  1,
				DefenderWin:  0,
				Draw:         0,
			},
		},
	}

	for i, tt := range values {
		result := tt.initialSet.Add(tt.addition)
		if !reflect.DeepEqual(result, tt.resultSet) {
			t.Errorf("Did not add a FirstRoundResult correctly to a collection\nExpected: %+v\nActual: %+v\nItem: %v", tt.resultSet, result, i)
		}
	}

}

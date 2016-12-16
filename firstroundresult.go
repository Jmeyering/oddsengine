package oddsengine

type FirstRoundResult struct {
	AttackerHits int `json:"attackerHits"`
	DefenderHits int `json:"defenderHits"`
	Frequency    int `json:"frequency"`
	AttackerWin  int `json:"attackerWin"`
	DefenderWin  int `json:"defenderWin"`
	Draw         int `json:"draw"`
}

type FirstRoundResultCollection []FirstRoundResult

func (fc FirstRoundResultCollection) Add(result FirstRoundResult) FirstRoundResultCollection {
	var added bool
	for i, f := range fc {
		added = false
		if f.AttackerHits == result.AttackerHits && f.DefenderHits == result.DefenderHits {
			fc[i].Frequency += result.Frequency
			fc[i].AttackerWin += result.AttackerWin
			fc[i].DefenderWin += result.DefenderWin
			fc[i].Draw += result.Draw
			added = true
			break
		}
	}
	if !added {
		fc = append(fc, result)
	}
	return fc
}

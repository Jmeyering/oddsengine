package oddsengine

import (
	"math"
	"strconv"
	"strings"
)

// Summary is a type which represents the an averaged results of multiple
// conflicts.
type Summary struct {
	// TotalSimulations The number of simulations that have been ran
	TotalSimulations int `json:"totalSimulations"`

	// AverageRounds The number of rounds on average a conflict lasted.
	AverageRounds float64 `json:"averageRounds"`

	// AttackerWinPercentage The percentage of conflicts that the attacker won
	AttackerWinPercentage float64 `json:"attackerWinPercentage"`

	// DefenderWinPercentage The percentage of conflicts that the defender won
	DefenderWinPercentage float64 `json:"defenderWinPercentage"`

	// DrawPercentage The percentage of conflicts that was a draw
	DrawPercentage float64 `json:"drawPercentage"`

	// AAAHitsAverage The number of AAA hits per round on average
	AAAHitsAverage float64 `json:"aaaHitsAverage"`

	// KamikazeHitsAverage The number of kamikaze hits per round on average
	KamikazeHitsAverage float64 `json:"kamikazeHitsAverage"`

	// AttackerAvgIpcLoss The number of IPC's the attacker loses on average
	AttackerAvgIpcLoss float64 `json:"attackerAvgIpcLoss"`

	// DefenderAvgIpcLoss The number of IPC's the defender loses on average
	DefenderAvgIpcLoss float64 `json:"defenderAvgIpcLoss"`

	// FirstRoundResults is the array of first round data. Represents the
	// number of hits that an attacker and defender get on the first round,
	// the frequency of such a result, and the victory result of that conflict.
	FirstRoundResults FirstRoundResultCollection `json:"firstRoundResults"`

	// AttackerUnitsRemaining represents all the remaining units at the end
	// of conflict. The units are represented by a string and the number of
	// times that that formation remained at the end of the conflict is the
	// value
	AttackerUnitsRemaining map[string]int `json:"attackerUnitsRemaining"`

	// DefenderUnitsRemaining represents all the remaining units at the end
	// of conflict. The units are represented by a string and the number of
	// times that that formation remained at the end of the conflict is the
	// value
	DefenderUnitsRemaining map[string]int `json:"defenderUnitsRemaining"`
}

// generateSummary Creates a summary from a slice of profiles.
func generateSummary(p []ConflictProfile) *Summary {

	var summary Summary
	summary.AttackerUnitsRemaining = map[string]int{}
	summary.DefenderUnitsRemaining = map[string]int{}
	summary.TotalSimulations = len(p)

	var totalRounds float64
	var totalAAAHits float64
	var totalKamikazeHits float64
	var totalAttackerWins float64
	var totalDefenderWins float64
	var totalDraw float64
	var totalAttackerIpcLoss float64
	var totalDefenderIpcLoss float64

	for _, profile := range p {

		if profile.Outcome == 0 {
			totalDraw++
		} else if profile.Outcome == 1 {
			totalAttackerWins++

			attackerRemainingString := formationSliceToString(profile.AttackerUnitsRemaining)

			if _, ok := summary.AttackerUnitsRemaining[attackerRemainingString]; ok {
				summary.AttackerUnitsRemaining[attackerRemainingString]++
			} else {
				summary.AttackerUnitsRemaining[attackerRemainingString] = 1
			}
		} else if profile.Outcome == -1 {
			totalDefenderWins++

			defenderRemainingString := formationSliceToString(profile.DefenderUnitsRemaining)

			if _, ok := summary.DefenderUnitsRemaining[defenderRemainingString]; ok {
				summary.DefenderUnitsRemaining[defenderRemainingString]++
			} else {
				summary.DefenderUnitsRemaining[defenderRemainingString] = 1
			}
		}

		firstRoundResult := FirstRoundResult{
			AttackerHits: profile.AttackerHits[0],
			DefenderHits: profile.DefenderHits[0],
			Frequency:    1,
		}

		if profile.Outcome == 0 {
			firstRoundResult.Draw = 1
		} else if profile.Outcome == 1 {
			firstRoundResult.AttackerWin = 1
		} else {
			firstRoundResult.DefenderWin = 1
		}
		summary.FirstRoundResults = summary.FirstRoundResults.Add(firstRoundResult)

		totalRounds += float64(profile.Rounds)
		totalAttackerIpcLoss += float64(profile.AttackerIpcLoss)
		totalDefenderIpcLoss += float64(profile.DefenderIpcLoss)
		totalAAAHits += float64(profile.AAAHits)
		totalKamikazeHits += float64(profile.KamikazeHits)
	}
	summary.AttackerWinPercentage = round((totalAttackerWins/float64(len(p)))*100, 2)
	summary.DefenderWinPercentage = round((totalDefenderWins/float64(len(p)))*100, 2)
	summary.DrawPercentage = round((totalDraw/float64(len(p)))*100, 2)
	summary.AttackerAvgIpcLoss = round((totalAttackerIpcLoss / float64(len(p))), 2)
	summary.AAAHitsAverage = round((totalAAAHits / float64(len(p))), 2)
	summary.KamikazeHitsAverage = round((totalKamikazeHits / float64(len(p))), 2)
	summary.DefenderAvgIpcLoss = round((totalDefenderIpcLoss / float64(len(p))), 2)
	summary.AverageRounds = round((totalRounds / float64(len(p))), 2)

	return &summary
}

func formationSliceToString(fs []map[string]int) string {
	var ss []string
	for _, f := range fs {
		for u, n := range f {
			ss = append(ss, u+":"+strconv.Itoa(n))
		}
	}
	return strings.Join(ss, ",")
}

// Round limits all floats to 2 decimal places
func round(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor((f*shift)+.5) / shift
}

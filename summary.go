package oddsengine

import "math"

// Summary is a type which represents the an averaged results of multiple
// conflicts.
type Summary struct {
	// AverageRounds The number of rounds on average a conflict lasted.
	AverageRounds float64

	// AttackerWinPercentage The percentage of conflicts that the attacker won
	AttackerWinPercentage float64

	// DefenderWinPercentage The percentage of conflicts that the defender won
	DefenderWinPercentage float64

	// DrawPercentage The percentage of conflicts that was a draw
	DrawPercentage float64

	// AAAHitsAverage The number of AAA hits per round on average
	AAAHitsAverage float64

	// KamikazeHitsAverage The number of kamikaze hits per round on average
	KamikazeHitsAverage float64

	// AttackerAvIpcLoss The number of IPC's the attacker loses on average
	AttackerAvIpcLoss float64

	// DefenderAvIpcLoss The number of IPC's the defender loses on average
	DefenderAvIpcLoss float64
}

// generateSummary Creates a summary from a slice of profiles.
func generateSummary(p []ConflictProfile) *Summary {

	var summary Summary
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
			totalDraw += 1
		} else if profile.Outcome == 1 {
			totalAttackerWins += 1
		} else if profile.Outcome == -1 {
			totalDefenderWins += 1
		}

		totalRounds += float64(profile.Rounds)
		totalAttackerIpcLoss += float64(profile.AttackerIpcLoss)
		totalDefenderIpcLoss += float64(profile.DefenderIpcLoss)
		totalAAAHits += float64(profile.AAAHits)
		totalKamikazeHits += float64(profile.KamikazeHits)
	}
	summary.AttackerWinPercentage = round((totalAttackerWins/float64(len(p)))*100, 2)
	summary.DefenderWinPercentage = round((totalDefenderWins/float64(len(p)))*100, 2)
	summary.DrawPercentage = round((totalDraw/float64(len(p)))*100, 2)
	summary.AttackerAvIpcLoss = round((totalAttackerIpcLoss / float64(len(p))), 2)
	summary.AAAHitsAverage = round((totalAAAHits / float64(len(p))), 2)
	summary.KamikazeHitsAverage = round((totalKamikazeHits / float64(len(p))), 2)
	summary.DefenderAvIpcLoss = round((totalDefenderIpcLoss / float64(len(p))), 2)
	summary.AverageRounds = round((totalRounds / float64(len(p))), 2)

	return &summary
}

// Round limits all floats to 2 decimal places
func round(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return math.Floor((f*shift)+.5) / shift
}

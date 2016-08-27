package oddsengine

// ConflictProfile is a struct representing the outcome of a single conflict
type ConflictProfile struct {
	// Rounds is the number of rounds a conflict took
	Rounds int

	// DefenderHits is number of losses by round for the attacker
	DefenderHits []int

	// AttackerHits is number of losses by round for the defender
	AttackerHits []int

	// The number of IPC's that the attacker lost in the conflict
	AttackerIpcLoss int

	// The number of IPC's that the defender lost in the conflict
	DefenderIpcLoss int

	// Number of Attacking Units Remaining at the end of the Conflict
	AttackerUnitsRemaining []map[string]int

	// Number of Defending Units Remaining at the end of the Conflict
	DefenderUnitsRemaining []map[string]int

	// AAA Hits represent the number of AAA hits for the conflict
	AAAHits int

	// AAA Hits represent the number of Kamikaze hits for the conflict
	KamikazeHits int

	// Outcome represents the status of the conflict after the fact.
	//  1: Attacker Victory
	//  0: Draw
	// -1: Defender Victory
	Outcome int
}

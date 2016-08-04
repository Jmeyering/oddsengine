package oddsengine

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

var (
	// mustTakeTerritory is a flag that will reserve the highest value land unit,
	// to allow the attacker to take the territory
	mustTakeTerritory bool

	// iterations is the number of times we will run the sim and generate a
	// ConflictProfile for the Summary. Default is 1000
	iterations = 1000

	// activePieces are the pieces available in the current game version
	// default pieces are 1940 pieces.
	activePieces = getPiecesForGame("1940")

	// activeGame is the game current being run by the simulator
	activeGame = "1940"

	// oolProfile is the general strategy for taking losses. Possible values are
	// "cost" and "hitValue"
	oolProfile = "cost"
)

// init the random seed for this run of the engine
func init() {
	rand.Seed(time.Now().UTC().UnixNano())
	setupOol()
}

// GetSummary is the function that ties everything together, Returns a summary
// of the conflict.
func GetSummary(attackers, defenders map[string]int) (*Summary, error) {
	var err error

	err = checkPieceValidity(attackers)
	if err != nil {
		return &Summary{}, err
	}

	err = checkPieceValidity(defenders)
	if err != nil {
		return &Summary{}, err
	}
	var profiles []ConflictProfile

	if mustTakeTerritory {
		reserveHighestValueLandUnit(attackers)
	}

	ool := customizeOol(attackers, defenders)
	for i := 0; i < iterations; i++ {
		profiles = append(profiles, *resolveConflict(attackers, defenders, ool))
	}
	return generateSummary(profiles), nil
}

// SetBaseOol allow a custom baseOol to be set for the conflict.
func SetBaseOol(ool []string) {
	baseOol = ool
}

// SetIterations changes the number of times the simulation will be ran.
func SetIterations(i int) {
	iterations = i
}

// SetMustTakeTerritory toggles the mustTakeTerritory flag for the simulation
func SetMustTakeTerritory(a bool) {
	mustTakeTerritory = a
}

// SetGame sets the game up internally. Altering piece makeup, and ool
func SetGame(g string) {
	activeGame = g
	activePieces = getPiecesForGame(g)
	setupOol()
}

// resolveConflict is the big boy here. When given a map of attacking and
// defending units, will resolve the conflict into a profile representing a
// whole host of data about the conflict.
func resolveConflict(a, d map[string]int, ool []string) *ConflictProfile {
	// We need to copy the passed in attackers and defenders so as to not
	// destroy the orininal map.
	attackers := make(map[string]int, len(a))
	defenders := make(map[string]int, len(d))

	// copy the attackers and defenders into new maps so we don't overwrite the
	// originals
	for k, v := range a {
		attackers[k] = v
	}
	for k, v := range d {
		defenders[k] = v
	}

	profile := new(ConflictProfile)

	// Let's loop infinitely here because we don't know how many rounds the
	// conflict will lets. And technically, the conflict CAN go on infinitely.
	for {

		// If the battle is resolved we can exit here
		if isResolved(attackers, defenders) {
			break
		}

		// Check if the defender has units capable of defending. If so we
		// calculate those casualties.
		if conflictIsAutoKill(defenders, attackers, len(profile.DefenderHits) == 0) {
			profile.DefenderIpcLoss += takeCasualties(defenders, getTotalNumUnits(defenders), ool)
			break
		}

		// If the battle is resolved we can exit here Yes we are checking again.
		// because the auto kill casualties may resolve the battle for us.
		if isResolved(attackers, defenders) {
			break
		}
		// We need to calculate the TOTAL number of hits by attacker and defender
		// for the round. This will be used in the profile and summary.
		// Also keep track of air + sub hits, because those must be allocated
		// separately
		var attackingHits int
		var attackingSubHits int
		var attackingAirHits int
		var defendingHits int
		var defendingSubHits int
		var defendingAirHits int

		/**
		 * Pre Conflict Attacks / Defence
		 *
		 * Perform any special pre-conflict attack or defence moves here.
		 * AAA guns, Kamikaze, Bombard etc.
		 */
		if len(profile.DefenderHits) == 0 {
			if canKamikaze(defenders) {

				// This is partially inaccurate, the kamikaze hits are limited by
				// the total number of hits available to be given on surface
				// ships MAX. To be completely accurate reallly, we need to accept
				// some form of input regarding which ships the kamikaze were
				// assigned to, however that isn't within the scope ATM.
				kamikazeRollMap := getKamikazeRollMap(defenders)
				kamikazeHits := calculateHits(kamikazeRollMap)
				profile.KamikazeHits = kamikazeHits
				if kamikazeHits > 0 {
					profile.AttackerIpcLoss += takeCasualties(attackers, kamikazeHits, surfaceShips)
				}

				// kamikaze are a one time use so delete them here.
				deleteUnitFromFormation(defenders, "kam")
			}

			// If we have AAA ability in the zone, we need to calculate those hits
			// first, and resolve the casualties before the defender is able to
			// fire back.
			if canUseAAA(attackers, defenders) {
				AAARollMap := getAAARollMap(attackers, defenders)
				AAAHits := calculateHits(AAARollMap)
				profile.AAAHits = AAAHits

				if AAAHits > 0 {
					profile.AttackerIpcLoss += takeCasualties(attackers, AAAHits, aircraft)
				}
			}

			// Ships that are capable of bombardment must go in this phase. They
			// do not prevent the hit defenders from attacking back, so we do
			// not take casualties.
			if canBombard(attackers) {
				bombardRollMap := getBombardRollMap(attackers)
				bombardHits := calculateHits(bombardRollMap)
				attackingHits += bombardHits

				// We need to remove the bombardships from the formation right
				// away to prevent them from getting hits assigned.
				for _, ship := range bombardShips {
					deleteUnitFromFormation(attackers, ship)
				}
			}
		}

		/**
		 * Submarine Suprise Attack
		 *
		 * Calculate the value of the submarine suprise attacks and take
		 * casualties accordingly
		 */

		// initialize there variables, we will need to track how many subs the
		// attacker and defender had before we take casualties
		var attackerNumSubs int
		var defenderNumSubs int
		var attackerSupriseHits int
		var attackerSubsLost int
		var defenderSupriseHits int
		var defenderSubsLost int

		// Calculate Submarine Suprise attacks
		attackerCanSuprise := canSupriseAttack(attackers, defenders)
		defenderCanSuprise := canSupriseAttack(defenders, attackers)

		// Defender and Attacker suprise attacks need to be calculated at the
		// same time. We aren't able to take casualties immediatly after,
		// because the defending submarine gets a shot no matter what, and we
		// don't want the attacking hit to destroy the sub, not allowing it to
		// get it's shot.
		if attackerCanSuprise {
			attackerNumSubs = attackers["sub"]
			rm := createRollMap(map[string]int{"sub": attackers["sub"]}, "attack")
			attackerSupriseHits = calculateHits(rm)
		}
		if defenderCanSuprise {
			defenderNumSubs = defenders["sub"]
			rm := createRollMap(map[string]int{"sub": defenders["sub"]}, "defend")
			defenderSupriseHits = calculateHits(rm)
		}

		// After the hits are calculated, we may take the casualties.
		if attackerSupriseHits > 0 {
			s := defenders["sub"]
			profile.DefenderIpcLoss += takeCasualties(defenders, attackerSupriseHits, ships)
			defenderSubsLost = s - defenders["sub"]
		}
		if defenderSupriseHits > 0 {
			s := attackers["sub"]
			profile.AttackerIpcLoss += takeCasualties(attackers, defenderSupriseHits, ships)
			attackerSubsLost = s - attackers["sub"]
		}

		/**
		 * Generate standard combat roll map
		 */
		attackerRollMap := createRollMap(attackers, "attack")
		defenderRollMap := createRollMap(defenders, "defend")

		/**
		 * Perform Roll Map Adjustments.
		 *
		 * We may need to reduce the number of shots that the roll map says,
		 * depending on if we have special units
		 */

		// If the defender has AAA We need to reduce the number of rolls at the
		// AAA hitValue
		if numAAA, ok := d["aaa"]; ok {
			defenderRollMap.Reduce(activePieces.Find("aaa").Defend, numAAA)
		}

		// We need to reduce the number of rolls in the roll map to account for
		// the subs that have already attacked.
		if attackerCanSuprise {
			reduction := attackerNumSubs - attackerSubsLost
			if reduction > 0 {
				attackerRollMap.Reduce(activePieces.Find("sub").Attack, reduction)
			}
		}
		if defenderCanSuprise {
			reduction := defenderNumSubs - defenderSubsLost
			if reduction > 0 {
				defenderRollMap.Reduce(activePieces.Find("sub").Defend, reduction)
			}
		}

		/**
		 * Calculate Roll Map Hits
		 *
		 * Using the roll map we get a "hit" score for both the attackers and
		 * defenders
		 */

		/**
		 * Roll Attackers First
		 */
		if hasLimitedAircraft(attackers, defenders) {
			rm := getAircraftRollMap(attackers, "attack")
			attackingAirHits = calculateHits(rm)
			for _, m := range rm {
				attackerRollMap.Reduce(m.hitValue, m.num)
			}
		}
		// We need to roll the subs separately from the other pieces, since they
		// cannot hit planes
		if numSub, ok := attackers["sub"]; ok && !attackerCanSuprise {
			rm := createRollMap(map[string]int{"sub": numSub}, "attack")
			attackingSubHits = calculateHits(rm)
			attackerRollMap.Reduce(activePieces.Find("sub").Attack, numSub)
		}
		// Calculate and record the attacking hits for the round.
		attackingHits += calculateHits(attackerRollMap)

		/**
		 * Roll Defenders Last
		 */

		if hasLimitedAircraft(defenders, attackers) {
			rm := getAircraftRollMap(defenders, "defend")
			defendingAirHits = calculateHits(rm)
			for _, m := range rm {
				defenderRollMap.Reduce(m.hitValue, m.num)
			}
		}

		if numSub, ok := defenders["sub"]; ok && !defenderCanSuprise {
			rm := createRollMap(map[string]int{"sub": numSub}, "defend")
			defendingSubHits = calculateHits(rm)
			defenderRollMap.Reduce(activePieces.Find("sub").Defend, numSub)
		}

		defendingHits += calculateHits(defenderRollMap)

		// Record data to the profile.
		profile.DefenderHits = append(profile.DefenderHits, defendingHits+defenderSupriseHits+defendingSubHits+defendingAirHits)
		profile.AttackerHits = append(profile.AttackerHits, attackingHits+attackerSupriseHits+attackingSubHits+attackingAirHits)

		/**
		 * Take Casualties
		 *
		 * Using the ool, we take casualties from the defenders, and the
		 * attackers
		 */

		// First take casualties from the submarines. Their hits can only be
		// applied to surface ships
		if attackingSubHits > 0 {
			profile.DefenderIpcLoss += takeCasualties(defenders, attackingSubHits, ships)
		}
		if defendingSubHits > 0 {
			profile.AttackerIpcLoss += takeCasualties(attackers, defendingSubHits, ships)
		}
		if attackingAirHits > 0 {
			profile.DefenderIpcLoss += takeCasualties(defenders, attackingAirHits, noSubOol)
		}
		if defendingAirHits > 0 {
			profile.AttackerIpcLoss += takeCasualties(attackers, defendingAirHits, noSubOol)
		}

		// Take standard combat casualties now.
		profile.DefenderIpcLoss += takeCasualties(defenders, attackingHits, ool)
		profile.AttackerIpcLoss += takeCasualties(attackers, defendingHits, ool)

	}

	// Record some more data to the profile
	profile.Rounds = len(profile.DefenderHits)

	if len(attackers) > 0 {
		profile.AttackerPiecesRemaining = formationToSortedSlice(attackers)
	}
	if len(defenders) > 0 {
		profile.DefenderPiecesRemaining = formationToSortedSlice(defenders)
	}

	// We record the conflict outcome onto the profile. Marked by
	//  1: Attacker Victory
	//  0: Draw
	// -1: Defender Victory
	if len(attackers) > 0 && len(defenders) > 0 {
		profile.Outcome = 0
	} else if len(attackers) == 0 && len(defenders) == 0 {
		profile.Outcome = 0
	} else if len(attackers) > 0 {
		profile.Outcome = 1
	} else if len(defenders) > 0 {
		profile.Outcome = -1
	}

	return profile

}

// rollDie functions as a random 1-6 number generator.
func rollDie() int {
	return rand.Intn(6) + 1
}

// multiRoll will roll a number of dice at a specific hitValue, returning the
// number of times the result of the die roll, was a hit according to the
// hitValue
func multiRoll(num, hitValue int) (hits int) {
	for i := 0; i < num; i++ {
		result := rollDie()
		if result <= hitValue {
			hits++
		}
	}

	return hits
}

// createRollMap will generate a RollMap from a given map of unit aliasas and
// number of them. Calculates the roll map with a given "mode", specifically,
// "attack" or "defend"
func createRollMap(p map[string]int, mode string) (rollMap RollMap) {
	rollMap = RollMap{}
	for piece, num := range p {
		if strings.HasPrefix(piece, "-") || strings.HasPrefix(piece, "+") {
			piece = piece[1:]
		}
		var hitValue int

		shotsAtPlusOne := 0
		pStruct := activePieces.Find(piece)

		hitValue = pStruct.Defend
		if mode == "attack" {
			hitValue = pStruct.Attack

			if pStruct.PlusOneShots != nil {
				shotsAtPlusOne = pStruct.PlusOneShots(p)
				num = num - shotsAtPlusOne
			}
		}

		if shotsAtPlusOne > 0 {
			rollMap = rollMap.AddRoll(hitValue+1, shotsAtPlusOne)
		}
		if num > 0 {
			rollMap = rollMap.AddRoll(hitValue, num)
		}
	}

	return rollMap
}

// calculateHits tallys the total number of hits for a map of units, returning
// the total number of hits.
func calculateHits(rollMap RollMap) (hits int) {
	for _, m := range rollMap {
		// If a map doesn't have a hit value, we don't need to roll for it.
		if m.hitValue == 0 {
			continue
		}

		hits += multiRoll(m.num, m.hitValue)
	}

	return hits
}

// getAircraftRollMap returns a roll map that encompases the aircraft only.
//
// @BUG because this method excludes ground troops, tactical bombers would not
// get their full slate of bonuses in certain circumstances. This is not as
// big of an issue however, because the primary purpose of this method, is to
// roll aircraft separately, when they canot hit submarines. So the scenario
// where we would want an aircraft exclusive rollmap will not involve tanks.
func getAircraftRollMap(a map[string]int, mode string) RollMap {
	manipulatedPlanes := make(map[string]int, len(aircraft))
	for _, plane := range aircraft {
		if _, ok := a[plane]; ok {
			manipulatedPlanes[plane] = a[plane]
		}
	}

	return createRollMap(manipulatedPlanes, mode)
}

// getBombardRollMap returns a rollmap of all available bombardable ships.
//
// Intentionally, we do not limit the number of bombard ships by any number of
// land units. We assume correct input. This is because it's impossible at this
// point to know which troops were offloaded via transport (The actual bombard
// limit) and which were brought in via a land territory. Incorrect input will
// lead to incorrect data.
func getBombardRollMap(a map[string]int) RollMap {
	manipulatedShips := make(map[string]int, len(bombardShips))

	for _, ship := range bombardShips {
		numShip := numAllUnitsInFormation(a, ship)
		if numShip > 0 {
			manipulatedShips[ship] = numShip
		}
	}

	return createRollMap(manipulatedShips, "attack")
}

// getKamikaze uses the defending units to calculate the number of
// rolls and the hit values that should be given to kamikaze
func getKamikazeRollMap(d map[string]int) RollMap {
	attackMap := make(map[string]int, 1)
	attackMap["kam"] = numAllUnitsInFormation(d, "kam")

	return createRollMap(attackMap, "defend")
}

// getAAARollMap uses the attackers and defenders to calculate the number of
// rolls that should be given to the AAA
func getAAARollMap(a, d map[string]int) RollMap {
	var numPlanes int
	numAAA := numAllUnitsInFormation(d, "aaa")

	// Determine how many planes the attacker has in it's fleet
	for _, plane := range aircraft {
		numPlanes += numAllUnitsInFormation(a, plane)
	}

	// By default the AAA can shoot off 3 shots per unit
	numAAAShots := numAAA * 3

	// but it is limited by the number of planes total
	if numAAAShots > numPlanes {
		numAAAShots = numPlanes
	}

	// create a "fake" unit map to calculate hits with
	manipulatedAAAPieces := map[string]int{
		"aaa": numAAAShots,
	}
	return createRollMap(manipulatedAAAPieces, "defend")
}

// takeCasualties removed units from the map in order of their value, and returns
// the total cost of the casualties taken.
func takeCasualties(units map[string]int, num int, ool []string) int {

	var ipcValueOfCasualties int
	// Find the units in order of their casualty value

	if hasUndamagedCapitalShips(units) {
		capitalShipDamage := damageCapitalShips(units, num)
		num = num - capitalShipDamage
	}

	for _, u := range ool {
		// If we are out of units to take, just stop.
		if num == 0 {
			break
		}

		// Depending on if the unit has a prefix or not, it's index may or may
		// not be the unit alias directly.
		unitIndex := u

		// The piece that we want to grab out of the "pieces" slice needs to be
		// recorded, this may change depending if we have a prefix.
		piecesIndex := u

		// If this piece is reserved, then the index within the pieces slice
		// is incorrect and we need to modify the lookup value to exclude the
		// "+"
		if strings.HasPrefix(unitIndex, "+") {
			piecesIndex = unitIndex[1:]
		}

		// Check for the existence of the unit in the map.
		numUnits, ok := units[unitIndex]
		if !ok {
			// The unit may be prefixed as a damaged prefix so check if that is
			// the case
			if numUnits, ok = units["-"+unitIndex]; ok {
				// So we have a damaged unit here update the unitIndex to
				// recognize that
				unitIndex = "-" + unitIndex
			} else {
				continue
			}
		}

		// Check the number of available units to take from If we can take the
		// entire unit set, then we do. Otherwise we remove as many as possible
		// from the unit set and reduce our number to 0.
		if numUnits <= num {
			num = num - numUnits
			ipcValueOfCasualties += (activePieces.Find(piecesIndex).Cost * numUnits)
			// Remove the unit from the unit set completely.
			delete(units, unitIndex)
		} else {
			ipcValueOfCasualties += (activePieces.Find(piecesIndex).Cost * num)
			units[unitIndex] = units[unitIndex] - num
			num = 0
		}
	}

	return ipcValueOfCasualties
}

// isResolved lets us know if the battle is over.
func isResolved(attackers, defenders map[string]int) (resolved bool) {
	if len(attackers) == 0 || len(defenders) == 0 {
		return true
	}

	_, defenderHasSub := defenders["sub"]

	if hasOnlyPlanes(attackers) && (len(defenders) == 1 && defenderHasSub) {
		return true
	}

	_, attackerHasSub := attackers["sub"]
	if hasOnlyPlanes(defenders) && (len(attackers) == 1 && attackerHasSub) {
		return true
	}

	return false
}

// damageCapitalShips assigns damage to capital, damage within the system is
// identified by a "-" before the alias name, for example. `bat` is an undamaged
// battleship. `-bat` is a damaged battleship. Returns the total number that
// were damaged.
func damageCapitalShips(units map[string]int, hits int) (numDamaged int) {
	for _, ship := range capitalShips {
		// If we don't have this capital ship, move on
		if _, ok := units[ship]; !ok {
			continue
		}

		// Record how many we already have damaged of this type
		numAlreadyDamaged, _ := units["-"+ship]

		numUnits := units[ship]

		// If we have enough hits to destroy all the undamaged capital ships, we
		// go ahead and take all the units
		if hits >= numUnits {
			delete(units, ship)
			units["-"+ship] = numUnits + numAlreadyDamaged
			numDamaged += numUnits
			hits = hits - numDamaged
		} else {
			units[ship] = units[ship] - hits
			units["-"+ship] = hits + numAlreadyDamaged
			numDamaged += hits
			break
		}
	}

	return numDamaged
}

// hasUndamagedCapitalShips will return whether or not a map of units has an
// undamaged capital ship.
func hasUndamagedCapitalShips(units map[string]int) bool {
	var a bool
	for _, ship := range capitalShips {
		// A rare circumstance where we ARE looking for an unprefixed unit.
		if _, ok := units[ship]; ok {
			a = true
			break
		}
	}

	return a
}

// canKamikaze returns whether or not the defender can kamikaze
func canKamikaze(units map[string]int) bool {
	return numAllUnitsInFormation(units, "kam") > 0
}

// attackerCanSupriseAttack lets us know if a conflict allows for a sub suprise
// attack by an attacker
func canSupriseAttack(a, b map[string]int) bool {
	aHasSub := numAllUnitsInFormation(a, "sub") > 0
	bHasDes := numAllUnitsInFormation(b, "des") > 0

	return aHasSub && !bHasDes
}

// canBombard lets us know if the units brought in allow for an offshore
// bombardment
func canBombard(units map[string]int) bool {

	return hasGroundUnits(units) && hasBombardShips(units)
}

// canUseAAA lets the program know if the current set of attackers and
// defenders are capable of using AAA before the start of the battle
func canUseAAA(attackers, defenders map[string]int) bool {
	_, ok := defenders["aaa"]

	return ok && hasAircraft(attackers)
}

// getTotalNumUnits returns the total number of units within a map of units
func getTotalNumUnits(u map[string]int) (num int) {
	for _, n := range u {
		num += n
	}
	return num
}

// conflictIsAutoKill returns whether or not the defender has any units
// capable of putting up a defensive hit. AAA is not a defending unit.
func conflictIsAutoKill(d, a map[string]int, firstRound bool) (autoKill bool) {
	autoKill = true
	for u := range d {
		if strings.HasPrefix(u, "-") || strings.HasPrefix(u, "+") {
			u = u[1:]
		}
		// The AAA is a special in that while it has a defend value, does
		// not actually get a defend shot in the combat phase.
		if u == "aaa" {
			// AAA only applies to the first round
			if !firstRound {
				continue
			}

			// If we have AAA and there are aircraft or there are no attackers
			// it is not an auto kill situation, otherwise aaa defenders is an
			// autokill
			if hasAircraft(a) || len(a) == 0 {
				autoKill = false
				break
			}

			continue
		}

		if activePieces.Find(u).Defend > 0 {
			autoKill = false
			break
		}
	}

	return autoKill
}

// reserveHighestValueLandUnit will assign the highest value unit in the units
// map as a reserved unit. Specifically by adding the "+" prefix to the unit
// alias. This unit will be taken last in conflict.
func reserveHighestValueLandUnit(units map[string]int) {
	// Iterate through the landTroops in reverse
	for i := len(landTroops) - 1; i >= 0; i-- {
		// If we have this land troop in our map, we need to reserve it.
		if _, ok := units[landTroops[i]]; ok {

			units[landTroops[i]] = units[landTroops[i]] - 1
			units["+"+landTroops[i]] = 1
			if units[landTroops[i]] == 0 {
				delete(units, landTroops[i])
			}
			break
		}
	}
}

// customizeOol takes the system's baseOol and customizes it for the particular
// pieces that have been passed in. The primary function in the real world is
// that it will add all reserved attackers and defenders to the appropriate
// spot in the ool, and add AAA to the end of the ool since, AAA must be taken
// last
func customizeOol(attackers, defenders map[string]int) []string {
	ool := make([]string, len(baseOol))
	copy(ool, baseOol)

	// We need to see all reserved attackers and add them to the end of the ool
	for alias := range attackers {
		if strings.HasPrefix(alias, "+") {
			ool = append(ool, alias)
		}
	}

	// We need to see all reserved defenders and add them to the end of the ool
	// Skipping those which have already been added
	for alias := range defenders {
		if strings.HasPrefix(alias, "+") {
			if !sliceHas(ool, alias) {
				ool = append(ool, alias)
			}
		}
	}

	// AAA Is always the last thing taken in any conflict
	if activePieces.HasPiece("aaa") {
		ool = append(ool, "aaa")
	}

	return ool
}

// sliceHas let's me know if a slice of strings has a particular value
func sliceHas(s []string, value string) bool {
	for _, a := range s {
		if a == value {
			return true
		}
	}
	return false
}

// checkPieceValidity determines if all the passed in pieces are valid for the
// particular game that is being simulated. If not valid, will return an error
// with a message including the pieces that are invalid.
func checkPieceValidity(p map[string]int) error {
	var invalid []string
	for alias := range p {
		if strings.HasPrefix(alias, "+") || strings.HasPrefix(alias, "-") {
			alias = alias[1:]
		}
		if activePieces.HasPiece(alias) {
			continue
		}

		invalid = append(invalid, alias)

	}

	if len(invalid) > 0 {
		return &InvalidPieceError{fmt.Sprintf("\"%v\"", strings.Join(invalid, "\", \""))}
	}

	return nil
}

// hasOnlyPlanes returns true if the formation contains only planes
func hasOnlyPlanes(u map[string]int) bool {
	for alias := range u {
		if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
			alias = alias[1:]
		}
		if has := sliceHas(aircraft, alias); !has {
			return false
		}
	}

	return true
}

// hasGroundUnits returns true if the formation contains any ground units
func hasGroundUnits(u map[string]int) bool {
	for alias := range u {
		if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
			alias = alias[1:]
		}
		if has := sliceHas(landTroops, alias); has {
			return true
		}
	}

	return false
}

// hasAircraft returns true if the formation contains any aircraft
func hasAircraft(u map[string]int) bool {
	for alias := range u {
		if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
			alias = alias[1:]
		}
		if has := sliceHas(aircraft, alias); has {
			return true
		}
	}

	return false
}

// hasBombardShips returns true if the formation contains any aircraft
func hasBombardShips(u map[string]int) bool {
	for alias := range u {
		if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
			alias = alias[1:]
		}
		if has := sliceHas(bombardShips, alias); has {
			return true
		}
	}

	return false
}

// deleteUnitFromFormation remove a unit and all its prefixed versions from a
// formation
func deleteUnitFromFormation(formation map[string]int, unit string) {
	delete(formation, unit)
	delete(formation, "+"+unit)
	delete(formation, "-"+unit)
}

// hasLimitedAircraft returns true if the first formation has aircraft which can
// not hit subs in the second formation.
func hasLimitedAircraft(a, b map[string]int) bool {
	var hasPlane bool
	for _, plane := range aircraft {
		if _, ok := a[plane]; ok {
			hasPlane = true
			break
		}
	}

	if !hasPlane {
		return false
	}
	if _, ok := b["sub"]; !ok {
		return false
	}
	if _, ok := a["des"]; ok {
		return false
	}
	return true
}

// formationToSortedSlice returns a unit formation as a slice that has been
// sorted according to it comparator. Because maps do not preserve order. When
// comparing map equality, or tranforming them to a slice, the resulting values
// may be different. If we instead transform the map into a slice, we can
// guarantee the correct order the units will come out as.
func formationToSortedSlice(f map[string]int) []map[string]int {
	ss := []map[string]int{}
	keys := make([]string, 0, len(f))
	for u := range f {
		keys = append(keys, u)
	}
	sort.Strings(keys)

	for _, k := range keys {
		ss = append(ss, map[string]int{k: f[k]})
	}

	return ss
}

// strictNumUnitsInFormation return the number of units matching a particular
// alias within the formation, Does NOT include damaged and reserved totals.
func strictNumUnitsInFormation(formation map[string]int, alias string) (num int) {
	num, _ = formation[alias]

	return num
}

// numAllUnitsInFormation return the TOTAL number of units matching a particular
// alias within the formation. Including damaged and reserved units.
func numAllUnitsInFormation(formation map[string]int, alias string) (num int) {
	num, _ = formation[alias]
	reservedNum, _ := formation["+"+alias]
	damagedNum, _ := formation["-"+alias]

	return num + reservedNum + damagedNum
}

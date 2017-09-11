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

	// activeGame is the game current being run by the simulator
	activeGame = "1940"

	// activeUnits are the units available in the current game version
	// default units are 1940 units.
	activeUnits = getUnitsForGame(activeGame)

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

	err = checkUnitValidity(attackers)
	if err != nil {
		return &Summary{}, err
	}

	err = checkUnitValidity(defenders)
	if err != nil {
		return &Summary{}, err
	}
	var profiles []ConflictProfile

	if mustTakeTerritory {
		reserveHighestValueLandUnit(attackers)
	}

	ool := customizeOol(attackers, defenders)
	ch := make(chan ConflictProfile, iterations)

	for i := 0; i < iterations; i++ {
		go func() {
			ch <- *resolveConflict(attackers, defenders, ool)
		}()
	}

	for i := 0; i < iterations; i++ {
		profiles = append(profiles, <-ch)
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

// SetGame sets the game up internally. Altering unit makeup, and ool
func SetGame(g string) {
	activeGame = g
	activeUnits = getUnitsForGame(g)
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
		var attackerAircraftHits int
		var defendingHits int
		var defendingSubHits int
		var defenderAircraftHits int

		/**
		 * Pre Conflict Attacks / Defence
		 *
		 * Perform any special pre-conflict attack or defence moves here.
		 * AAA guns, Kamikaze, Bombard etc.
		 */
		if len(profile.DefenderHits) == 0 {
			/**
			 * Run Kamikaze attacks.
			 */
			// This is partially inaccurate, the kamikaze hits are limited by
			// the total number of hits available to be given on surface
			// ships MAX. To be completely accurate reallly, we need to accept
			// some form of input regarding which ships the kamikaze were
			// assigned to, however that isn't within the scope ATM.
			kamikazeHits := rollForUnitSlice(defenders, []string{"kam"}, "defend")
			profile.KamikazeHits = kamikazeHits

			if kamikazeHits > 0 {
				profile.AttackerIpcLoss += takeCasualties(attackers, kamikazeHits, surfaceShips)
			}

			// kamikaze are a one time use so delete them here.
			deleteUnitFromFormation(defenders, "kam")

			/**
			 * Run AAA Attacks
			 */
			// If we have AAA ability in the zone, we need to calculate those hits
			// first, and resolve the casualties before the defender is able to
			// fire back.
			AAARollMap := getAAARollMap(attackers, defenders)
			AAAHits := calculateHits(AAARollMap)
			profile.AAAHits = AAAHits

			if AAAHits > 0 {
				profile.AttackerIpcLoss += takeCasualties(attackers, AAAHits, aircraft)
			}

			// Ships that are capable of bombardment must go in this phase. They
			// do not prevent the hit defenders from attacking back, so we do
			// not take casualties.
			if canBombard(attackers) {
				attackingHits += rollForUnitSlice(attackers, bombardShips, "attack")

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
		var attackerSupriseHits int
		var defenderSupriseHits int

		// Calculate Submarine Suprise attacks
		attackerCanSuprise := canSupriseAttack(attackers, defenders)
		defenderCanSuprise := canSupriseAttack(defenders, attackers)

		// Defender and Attacker suprise attacks need to be calculated at the
		// same time. We aren't able to take casualties immediatly after,
		// because the defending submarine gets a shot no matter what, and we
		// don't want the attacking hit to destroy the sub, not allowing it to
		// get it's shot.
		if attackerCanSuprise {
			attackerSupriseHits = rollSubs(attackers, "attack")
		}
		if defenderCanSuprise {
			defenderSupriseHits = rollSubs(defenders, "defend")
		}

		// After the hits are calculated, we may take the casualties.
		profile.DefenderIpcLoss += takeCasualties(defenders, attackerSupriseHits, ships)
		profile.AttackerIpcLoss += takeCasualties(attackers, defenderSupriseHits, ships)

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

		// Reduce the number of rolls at the AAA hitValue
		defenderRollMap.RemoveUnits(defenders, []string{"aaa", "raaa", "aag"}, "defend")

		// We need to reduce the number of rolls in the roll map to account for
		// the subs that have already attacked.
		if attackerCanSuprise {
			attackerRollMap.RemoveUnits(attackers, subs, "attack")
		}
		if defenderCanSuprise {
			defenderRollMap.RemoveUnits(defenders, subs, "defend")
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

		// Roll for units who's hits may be limited. Aircraft first.

		// Aircraft should only roll if there are units that they are able to
		// hit
		if !hasOnlySubs(defenders) || hasUnit(attackers, "des") {
			attackerAircraftHits = rollAircraft(attackers, "attack")
		}

		// Remove the aircraft from the roll map so we don't roll for them in
		// the later stages
		attackerRollMap.RemoveUnits(attackers, aircraft, "attack")

		// We need to know what the aircraft are able to hit. If they are limited
		// they are unable to hit submarines
		attackingAircraftOol := ool
		if hasLimitedAircraft(attackers, defenders) {
			attackingAircraftOol = noSubOol
		}

		// We need to roll the subs separately from the other units, since they
		// cannot hit planes
		if hasSub(attackers) && !attackerCanSuprise {
			attackingSubHits = rollSubs(attackers, "attack")
			attackerRollMap.RemoveUnits(attackers, subs, "attack")
		}

		// Calculate and record the attacking hits for the round.
		attackingHits += calculateHits(attackerRollMap)

		/**
		 * Roll Defenders Last
		 */

		// Roll for units who's hits may be limited. Aircraft first.

		// Aircraft should only roll if there are units that they are able to
		// hit
		if !hasOnlySubs(attackers) || hasUnit(defenders, "des") {
			defenderAircraftHits = rollAircraft(defenders, "defend")
		}

		// Remove the aircraft from the roll map so we don't roll for them twice
		defenderRollMap.RemoveUnits(defenders, aircraft, "defend")

		// We need to know what the aircraft are able to hit. If they are limited
		// they are unable to hit submarines
		defendingAircraftOol := ool
		if hasLimitedAircraft(defenders, attackers) {
			defendingAircraftOol = noSubOol
		}

		if hasSub(defenders) && !defenderCanSuprise {
			defendingSubHits = rollSubs(defenders, "defend")
			defenderRollMap.RemoveUnits(defenders, subs, "defend")
		}

		defendingHits += calculateHits(defenderRollMap)

		totalDefenderHits := defendingHits + defenderSupriseHits + defendingSubHits + defenderAircraftHits
		totalAttackerHits := attackingHits + attackerSupriseHits + attackingSubHits + attackerAircraftHits

		// Record data to the profile.
		profile.DefenderHits = append(profile.DefenderHits, totalDefenderHits)
		profile.AttackerHits = append(profile.AttackerHits, totalAttackerHits)

		/**
		 * Take Casualties
		 *
		 * Using the ool, we take casualties from the defenders, and the
		 * attackers
		 */

		// First take casualties from the submarines. Their hits can only be
		// applied to surface ships
		profile.DefenderIpcLoss += takeCasualties(defenders, attackingSubHits, ships) +
			takeCasualties(defenders, attackerAircraftHits, attackingAircraftOol) +
			takeCasualties(defenders, attackingHits, ool)

		profile.AttackerIpcLoss += takeCasualties(attackers, defendingSubHits, ships) +
			takeCasualties(attackers, defenderAircraftHits, defendingAircraftOol) +
			takeCasualties(attackers, defendingHits, ool)

	}

	// Record some more data to the profile
	profile.Rounds = len(profile.DefenderHits)

	if len(attackers) > 0 {
		profile.AttackerUnitsRemaining = formationToSortedSlice(attackers)
	}
	if len(defenders) > 0 {
		profile.DefenderUnitsRemaining = formationToSortedSlice(defenders)
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

// rollDie functions as a random number generator Rolls at 6 normally, but
// deluxe rolls an 8 sided die. This needs a good refactor.
func rollDie() int {
	rollBase := 6
	if activeGame == "deluxe" {
		rollBase = 8
	}
	return rand.Intn(rollBase) + 1
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
func createRollMap(f map[string]int, mode string) (rollMap RollMap) {
	for alias, n := range f {
		var hitValue int

		shotsAtPlusOne := 0
		totalNumUnits := numAllUnitsInFormation(f, realAlias(alias))

		hasModifiedUnits := totalNumUnits > n
		isModifiedUnit := strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+")

		if hasModifiedUnits && isModifiedUnit {
			continue
		}

		unit := activeUnits.Find(realAlias(alias))

		if mode == "attack" {
			hitValue = unit.Attack

			if unit.PlusOneRolls != nil {
				shotsAtPlusOne = unit.PlusOneRolls(f)
				totalNumUnits = totalNumUnits - shotsAtPlusOne
			}
		} else {
			hitValue = unit.Defend
			if unit.PlusOneDefend != nil {
				shotsAtPlusOne = unit.PlusOneDefend(f)
				totalNumUnits = totalNumUnits - shotsAtPlusOne
			}
		}

		if shotsAtPlusOne > 0 {
			rollMap = rollMap.AddRoll(hitValue+1, shotsAtPlusOne)
		}
		if totalNumUnits > 0 {
			rollMap = rollMap.AddRoll(hitValue, totalNumUnits)
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

// getAAARollMap uses the attackers and defenders to calculate the number of
// rolls that should be given to the AAA
func getAAARollMap(a, d map[string]int) RollMap {
	var numPlanes int

	// We need to determine if we are rolling for standard AAA or if we have
	// Radar assisted AAA guns
	unitAlias := "aaa"
	if _, hasRaaa := d["raaa"]; hasRaaa {
		unitAlias = "raaa"
	}

	// Refactor some day this is awful
	if activeGame == "deluxe" {
		unitAlias = "aag"
	}

	numAAA := numAllUnitsInFormation(d, unitAlias)

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
	aaaFormation := map[string]int{
		unitAlias: numAAAShots,
	}

	return createRollMap(aaaFormation, "defend")
}

// takeCasualties removed units from the map in order of their value, and returns
// the total cost of the casualties taken.
func takeCasualties(f map[string]int, num int, ool []string) int {
	// If we get a casualty number of 0 just leave that shit alone.
	if num <= 0 {
		return 0
	}

	var ipcValueOfCasualties int
	// Find the units in order of their casualty value

	if hasUndamagedCapitalShips(f) {
		capitalShipDamage := damageCapitalShips(f, num)
		num = num - capitalShipDamage
	}

	for _, u := range ool {
		// If we are out of units to take, just stop.
		if num == 0 {
			break
		}

		// The unitIndex is the index within the passed in unit map.
		// Depending on if the unit has a prefix or not, it's index may or may
		// not be the unit alias directly.
		unitIndex := u

		// The unmodifiedIndex is the index within the total game units.
		// The unit that we want to grab out of the "units" slice needs to be
		// recorded, this may change depending if we have a prefix.
		unmodifiedIndex := u

		// If this unit is reserved, then the index within the units slice
		// is incorrect and we need to modify the lookup value to exclude the
		// "+"
		if strings.HasPrefix(unitIndex, "+") {
			unmodifiedIndex = unitIndex[1:]
		}

		// Check for the existence of the unit in the map.
		numUnits, ok := f[unitIndex]
		if !ok {
			// The unit may be prefixed as a damaged prefix so check if that is
			// the case
			if numUnits, ok = f["-"+unitIndex]; ok {
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
			ipcValueOfCasualties += (activeUnits.Find(unmodifiedIndex).Cost * numUnits)
			// Remove the unit from the unit set completely.
			delete(f, unitIndex)
		} else {
			ipcValueOfCasualties += (activeUnits.Find(unmodifiedIndex).Cost * num)
			f[unitIndex] = f[unitIndex] - num
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

	defenderHasSub := hasSub(defenders)

	if hasOnlyPlanes(attackers) && (len(defenders) == 1 && defenderHasSub) {
		return true
	}

	attackerHasSub := hasSub(attackers)
	if hasOnlyPlanes(defenders) && (len(attackers) == 1 && attackerHasSub) {
		return true
	}

	return false
}

// damageCapitalShips assigns damage to capital, damage within the system is
// identified by a "-" before the alias name, for example. `bat` is an undamaged
// battleship. `-bat` is a damaged battleship. Returns the total number that
// were damaged.
// @TODO There is an error in here. We need to be able to assign damage to a
// reserved capital ship. How do we handle a `-+bat` or a `+-bat`
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

		// Find an undamaged capital ship. Can be reserved But can NOT be
		// damaged obvs. So don't use the HasUnits method here
		_, hasReserved := units["+"+ship]
		_, hasStandard := units[ship]
		if hasReserved || hasStandard {
			a = true
			break
		}
	}

	return a
}

// canKamikaze returns whether or not the defender can kamikaze
func canKamikaze(units map[string]int) bool {
	return hasUnit(units, "kam")
}

// attackerCanSupriseAttack lets us know if a conflict allows for a sub suprise
// attack by an attacker
func canSupriseAttack(a, b map[string]int) bool {
	aHasSub := hasSub(a)
	bHasDes := hasUnit(b, "des")

	return aHasSub && !bHasDes
}

// canBombard lets us know if the units brought in allow for an offshore
// bombardment. There is an issue here, if an end user sends through a ship as
// a bombard against a land unit, the conflict will proceed like a normal
// conflict. I'm calling this, however, "not a bug" but a user error.
func canBombard(units map[string]int) bool {

	return hasGroundUnits(units) && hasBombardShips(units)
}

// canUseAAA lets the program know if the current set of attackers and
// defenders are capable of using AAA before the start of the battle
func canUseAAA(attackers, defenders map[string]int) bool {
	return (hasUnit(defenders, "aaa") || hasUnit(defenders, "raaa") || hasUnit(defenders, "aag")) && hasAircraft(attackers)
}

// getTotalNumUnits returns the total number of units within a map of units
func getTotalNumUnits(u map[string]int) (num int) {
	for _, n := range u {
		num += n
	}
	return num
}

// rollForUnit rolls all the units identified by a particular alias and returns
// the number of hits.
func rollForUnit(f map[string]int, unit *Unit, mode string) (hits int) {
	numUnits := numAllUnitsInFormation(f, unit.Alias)

	if mode == "attack" {
		if unit.MultiRoll > 0 {
			hits = rollMultiRollUnits(map[string]int{unit.Alias: numUnits}, mode)
		} else {
			rm := RollMap{}
			var unitsAtPlusOne int
			if unit.PlusOneRolls != nil {
				unitsAtPlusOne = unit.PlusOneRolls(f)
				numUnits = numUnits - unitsAtPlusOne
			}

			if unitsAtPlusOne > 0 {
				rm = rm.AddRoll(unit.Attack+1, unitsAtPlusOne)
			}
			if numUnits > 0 {
				rm = rm.AddRoll(unit.Attack, numUnits)
			}
			hits = calculateHits(rm)
		}
	} else {
		rm := createRollMap(map[string]int{unit.Alias: numUnits}, mode)
		hits = calculateHits(rm)
	}

	return hits
}

func rollForUnitSlice(f map[string]int, slice []string, mode string) (hits int) {
	for _, alias := range slice {
		if hasUnit(f, alias) {
			unit := activeUnits.Find(realAlias(alias))
			hits += rollForUnit(f, unit, mode)
		}
	}

	return hits
}

// rollSubs is a convenienve method that rolls all the sub units and returns
// the number of hits
func rollSubs(a map[string]int, mode string) (hits int) {
	return rollForUnitSlice(a, subs, mode)
}

// rollAircraft is a convenience method that rolls all the aircraft units and
// returns the number of hits
func rollAircraft(a map[string]int, mode string) (hits int) {
	return rollForUnitSlice(a, aircraft, mode)
}

// rollMultiRollUnits rolls for all the units who get multiple dice per attack
// roll. Selecting the highest of the die to score a hit.
func rollMultiRollUnits(a map[string]int, mode string) (hits int) {
	for _, alias := range multiRollUnits {
		if hasUnit(a, alias) {
			// We must run each of these units separately to keep track
			// of their rolls. So we will iterate as many units as we have.
			iterations := numAllUnitsInFormation(a, alias)
			numDie := activeUnits.Find(realAlias(alias)).MultiRoll
			for i := 0; i < iterations; i++ {
				// Maybe a little bit of a hack. create a roll map for each
				// iteration through. if any hits come back, record just 1 hit.
				// since we are rolling multiple die but for only one unit.
				rm := createRollMap(map[string]int{alias: numDie}, mode)
				h := calculateHits(rm)
				if h > 0 {
					hits++
				}
			}
		}
	}

	return hits
}

// conflictIsAutoKill returns whether or not the defender has any units
// capable of putting up a defensive hit. AAA is not a defending unit.
func conflictIsAutoKill(d, a map[string]int, firstRound bool) (autoKill bool) {
	autoKill = true
	for alias := range d {
		alias = realAlias(alias)
		// The AAA is a special in that while it has a defend value, does
		// not actually get a defend shot in the combat phase.
		if alias == "aaa" || alias == "raaa" {
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

		if activeUnits.Find(alias).Defend > 0 {
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

// sliceHasUnit let's me know if a slice of strings has a particular value
func sliceHasUnit(s []string, alias string) bool {
	if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
		alias = alias[1:]
	}
	for _, a := range s {
		if a == alias {
			return true
		}
	}
	return false
}

// sliceHasValue let's me know if a slice of strings has a particular value,
// does not check for modifiers
func sliceHasValue(s []string, alias string) bool {
	for _, a := range s {
		if a == alias {
			return true
		}
	}
	return false
}

// checkUnitValidity determines if all the passed in units are valid for the
// particular game that is being simulated. If not valid, will return an error
// with a message including the units that are invalid.
func checkUnitValidity(p map[string]int) error {
	var invalid []string
	for alias := range p {
		if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
			alias = alias[1:]
		}
		if !activeUnits.HasUnit(alias) {
			invalid = append(invalid, alias)
		}
	}

	if len(invalid) > 0 {
		return &InvalidUnitError{fmt.Sprintf("Invalid Unit(s) supplied:\n%s", strings.Join(invalid, ", "))}
	}

	return nil
}

// hasOnlyPlanes returns true if the formation contains only planes
func hasOnlyPlanes(u map[string]int) bool {
	for alias := range u {
		if has := sliceHasUnit(aircraft, alias); !has {
			return false
		}
	}

	return true
}

// hasOnlySubs returns true if the formation contains only subs
func hasOnlySubs(u map[string]int) bool {
	for alias := range u {
		if has := sliceHasUnit(subs, alias); !has {
			return false
		}
	}

	return true
}

// hasGroundUnits returns true if the formation contains any ground units
func hasGroundUnits(u map[string]int) bool {
	for _, unit := range landTroops {
		if hasUnit(u, unit) {
			return true
		}
	}
	return false
}

// hasSub returns true if the formation contains any submarines
func hasSub(u map[string]int) bool {
	for _, unit := range subs {
		if hasUnit(u, unit) {
			return true
		}
	}
	return false
}

// hasUnit determines if a unit exists in a formation. The unit may be damaged
// or reserved and still return true.
func hasUnit(units map[string]int, alias string) bool {
	_, has := units[alias]
	_, hasReserved := units["+"+alias]
	_, hasDamaged := units["-"+alias]

	return has || hasReserved || hasDamaged

}

// hasAircraft returns true if the formation contains any aircraft
func hasAircraft(u map[string]int) bool {
	for _, unit := range aircraft {
		if hasUnit(u, unit) {
			return true
		}
	}
	return false
}

// hasBombardShips returns true if the formation contains any aircraft
func hasBombardShips(u map[string]int) bool {
	for _, unit := range bombardShips {
		if hasUnit(u, unit) {
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

	if !hasAircraft(a) {
		return false
	}

	if !hasSub(b) {
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

// numAllUnitsInFormation return the TOTAL number of units matching a particular
// alias within the formation. Including damaged and reserved units.
func numAllUnitsInFormation(formation map[string]int, alias string) (num int) {
	num, _ = formation[alias]
	reservedNum, _ := formation["+"+alias]
	damagedNum, _ := formation["-"+alias]

	return num + reservedNum + damagedNum
}

// realAlias returns the actual alias of a unit. Trimming any modifiers
func realAlias(alias string) string {
	if strings.HasPrefix(alias, "-") || strings.HasPrefix(alias, "+") {
		alias = alias[1:]
	}

	return alias
}

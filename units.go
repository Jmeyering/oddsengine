package oddsengine

// Unit represents a specific unit within a game, identifing a lot of specific
// information related to the unit.
type Unit struct {
	Alias            string
	Name             string
	Cost             int
	Attack           int
	Defend           int
	IsShip           bool
	IsSub            bool
	IsAircraft       bool
	CapitalShip      bool
	CanBombard       bool
	CanTakeTerritory bool
	PlusOneRolls     func(map[string]int) int
	// MultiRoll is the number of dice the unit can roll, and select the best
	// roll for it's hit.
	MultiRoll int
}

// Units is a container for multiple Unit structs
type Units []Unit

// Swap implementing Sortable
func (p Units) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Len implementing Sortable
func (p Units) Len() int {
	return len(p)
}

// Create three different Sortable containers for our Units

// ByDefendingPower sorts the units by the higest Defend value of the unit
type ByDefendingPower struct{ Units }

// ByAttackingPower sorts the units by the higest Attack value of the unit
type ByAttackingPower struct{ Units }

// ByCost sorts the units by the lowest Cost value of the unit
type ByCost struct{ Units }

// Less implementing Sortable
func (p ByDefendingPower) Less(i, j int) bool {
	if p.Units[i].Alias == "aaa" {
		return false
	} else if p.Units[j].Alias == "aaa" {
		return true
	} else if p.Units[i].Defend == p.Units[j].Defend {
		return p.Units[i].Cost < p.Units[j].Cost
	}

	return p.Units[i].Defend < p.Units[j].Defend
}

// Less implementing Sortable
func (p ByAttackingPower) Less(i, j int) bool {
	if p.Units[i].Alias == "aaa" {
		return false
	} else if p.Units[j].Alias == "aaa" {
		return true
	} else if p.Units[i].Attack == p.Units[j].Attack {
		return p.Units[i].Cost < p.Units[j].Cost
	}

	return p.Units[i].Attack < p.Units[j].Attack
}

// Less implementing Sortable
func (p ByCost) Less(i, j int) bool {
	if p.Units[i].Alias == "aaa" {
		return false
	} else if p.Units[j].Alias == "aaa" {
		return true
	} else if p.Units[i].Cost == p.Units[j].Cost {
		return p.Units[i].Attack < p.Units[j].Attack
	}

	return p.Units[i].Cost < p.Units[j].Cost
}

// getUnitsForGame returns a Units type containing all the units that are
// valid for a particular game identified by the game string passed in
func getUnitsForGame(game string) (p Units) {
	switch game {
	case "1940":
		p = get1940Units()
	case "1941":
		p = get1941Units()
	case "1942":
		p = get1942Units()
	}
	return p
}

// Delete removes an item from a units type and returns the new type
func (p Units) Delete(alias string) Units {
	for i, unit := range p {
		if unit.Alias == alias {
			copy(p[i:], p[i+1:])
			p[len(p)-1] = Unit{}
			p = p[:len(p)-1]
		}
	}

	return p
}

// HasUnit tells us if a Units type has a unit identified by the alias
// provided
func (p Units) HasUnit(alias string) (has bool) {
	for _, unit := range p {
		if unit.Alias == alias {
			has = true
			break
		}
	}

	return has
}

// Find a specific unit, by alias, within the slice of Units
func (p Units) Find(alias string) *Unit {
	var targetUnit Unit
	for _, unit := range p {
		if unit.Alias == alias {
			targetUnit = unit
			break
		}
	}

	return &targetUnit
}

// get1940Units returns the units that are valid for the game Axis and Allies
// 1940
func get1940Units() Units {
	p := get1942Units()

	// The carrier in 1940 is now a capital ship
	p = p.Delete("car")

	p = append(p,
		// While not really a "unit" per se, we will treat it like one. It
		// should be deleted from any unit mapping once it is used
		Unit{
			Alias:  "kam",
			Name:   "Kamikaze",
			Cost:   0,
			Attack: 0,
			Defend: 2,
		},
		Unit{
			Alias:  "mec",
			Name:   "Mechanized Infantry",
			Cost:   4,
			Attack: 1,
			Defend: 2,
			// The mec defers all it's +1 shots to the inf, only taking whatever
			// shots are available
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var remainingShots int

				var pairedArtilleryShotsAvailable int

				numInf := numAllUnitsInFormation(u, "inf")
				numMec := numAllUnitsInFormation(u, "mec")
				numArt := numAllUnitsInFormation(u, "art")
				numAArt := numAllUnitsInFormation(u, "aart")

				pairedArtilleryShotsAvailable = numArt + (numAArt * 2)

				if pairedArtilleryShotsAvailable == 0 {
					return shots
				}

				// 2 mec 3 art
				if numInf > 0 {
					remainingShots = pairedArtilleryShotsAvailable - numInf
				} else {
					remainingShots = pairedArtilleryShotsAvailable
				}

				if remainingShots <= 0 {
					return shots
				}

				if remainingShots < numMec {
					shots = remainingShots
				} else {
					shots = numMec
				}

				return shots
			},
			CanTakeTerritory: true,
		},
		Unit{
			Alias:      "tac",
			Name:       "Tactical Bomber",
			Cost:       11,
			Attack:     3,
			Defend:     3,
			IsAircraft: true,
			PlusOneRolls: func(u map[string]int) int {
				var shots int

				numTac := numAllUnitsInFormation(u, "tac")
				numFig := numAllUnitsInFormation(u, "fig")
				numTan := numAllUnitsInFormation(u, "tan")

				totalNumBoosters := numFig + numTan

				if totalNumBoosters == 0 {
					return shots
				}

				// The number of shots are limited by the total number of tac
				// within the unit group
				shots = numTac
				if totalNumBoosters < numTac {
					shots = totalNumBoosters
				}

				return shots
			},
		},
		Unit{
			Alias:       "car",
			Name:        "Aircraft Carrier",
			Cost:        16,
			Attack:      0,
			Defend:      2,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: true,
		},
		Unit{
			Alias:      "hbom",
			Name:       "Heavy Bomber",
			Cost:       12,
			Attack:     4,
			Defend:     1,
			IsAircraft: true,
			MultiRoll:  2,
		},
		Unit{
			Alias:  "raaa",
			Name:   "Anti-Aircraft Artillery",
			Cost:   5,
			Attack: 0,
			Defend: 2,
		},
		Unit{
			Alias:      "jfig",
			Name:       "Jet Fighters",
			Cost:       10,
			IsAircraft: true,
			Attack:     4,
			Defend:     4,
		},
		Unit{
			Alias:  "ssub",
			Name:   "Super Sumbarine",
			IsShip: true,
			IsSub:  true,
			Cost:   6,
			Attack: 3,
			Defend: 1,
		},
		Unit{
			Alias:  "imec",
			Name:   "Mechanized Infantry",
			Cost:   4,
			Attack: 1,
			Defend: 2,
			// The imec defers all it's +1 shots to the inf, only taking
			// whatever shots are available or pairs with a tank for shots
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var remainingShots int
				var pairedArtilleryShotsAvailable int

				numInf := numAllUnitsInFormation(u, "inf")
				numTan := numAllUnitsInFormation(u, "tan")
				numArt := numAllUnitsInFormation(u, "art")
				numAArt := numAllUnitsInFormation(u, "aart")
				numMec := numAllUnitsInFormation(u, "imec")

				pairedArtilleryShotsAvailable = numArt + (numAArt * 2)
				remainingShots = pairedArtilleryShotsAvailable - numInf

				if remainingShots <= 0 {
					remainingShots = numTan
				} else {
					remainingShots += numTan
				}

				if remainingShots <= 0 {
					return shots
				}

				if remainingShots < numMec {
					shots = remainingShots
				} else {
					shots = numMec
				}

				return shots
			},
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "aart",
			Name:             "Advanced Artillery",
			Cost:             4,
			Attack:           2,
			Defend:           2,
			CanTakeTerritory: true,
		},
	)

	return p
}

// get1942Units returns the units that are valid for the game Axis and Allies
// 1942
func get1942Units() Units {
	p := get1941Units()
	// The 1942 Battleship gains the ability to bombard, it is also more
	// expensive
	p = p.Delete("bat")

	// The 1942 Battleship is more expensive
	p = p.Delete("car")
	p = append(p,
		Unit{
			Alias:  "aaa",
			Name:   "Anti-Aircraft Artillery",
			Cost:   5,
			Attack: 0,
			Defend: 1,
		},
		Unit{
			Alias:            "art",
			Name:             "Artillery",
			Cost:             4,
			Attack:           2,
			Defend:           2,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:      "cru",
			Name:       "Cruiser",
			Cost:       12,
			Attack:     3,
			Defend:     3,
			IsShip:     true,
			CanBombard: true,
		},
		Unit{
			Alias:       "bat",
			Name:        "Battleship",
			Cost:        20,
			Attack:      4,
			Defend:      4,
			IsShip:      true,
			CanBombard:  true,
			CapitalShip: true,
		},
		Unit{
			Alias:       "car",
			Name:        "Aircraft Carrier",
			Cost:        14,
			Attack:      1,
			Defend:      2,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: false,
		},
	)

	return p
}

// get1941Units returns the units that are valid for the game Axis and Allies
// 1941
func get1941Units() Units {
	return Units{
		Unit{
			Alias:  "inf",
			Name:   "Infantry",
			Cost:   3,
			Attack: 1,
			Defend: 2,
			// Even though in 1941, there is no +1 for INF, having this here
			// does no harm because the conditions will never be met
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var pairedArtilleryShotsAvailable int

				numInf := numAllUnitsInFormation(u, "inf")
				numArt := numAllUnitsInFormation(u, "art")
				numAArt := numAllUnitsInFormation(u, "aart")

				pairedArtilleryShotsAvailable = numArt + (numAArt * 2)

				if pairedArtilleryShotsAvailable == 0 {
					return shots
				}

				// Assume they will all be paired
				shots = numInf

				// If they can't all be paired, return the total number of
				// possible pairings
				if pairedArtilleryShotsAvailable < numInf {
					shots = pairedArtilleryShotsAvailable
				}

				return shots
			},
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "tan",
			Name:             "Tank",
			Cost:             6,
			Attack:           3,
			Defend:           3,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:      "fig",
			Name:       "Fighter",
			Cost:       10,
			Attack:     3,
			Defend:     4,
			IsAircraft: true,
		},
		Unit{
			Alias:      "bom",
			Name:       "Strategic Bomber",
			Cost:       12,
			Attack:     4,
			Defend:     1,
			IsAircraft: true,
		},
		Unit{
			Alias:  "sub",
			Name:   "Submarine",
			Cost:   6,
			IsShip: true,
			IsSub:  true,
			Attack: 2,
			Defend: 1,
		},
		Unit{
			Alias:  "des",
			Name:   "Destroyer",
			Cost:   8,
			Attack: 2,
			Defend: 2,
			IsShip: true,
		},
		Unit{
			Alias:       "car",
			Name:        "Aircraft Carrier",
			Cost:        12,
			Attack:      1,
			Defend:      2,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: false,
		},
		Unit{
			Alias:       "bat",
			Name:        "Battleship",
			Cost:        16,
			Attack:      4,
			Defend:      4,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: true,
		},
	}
}

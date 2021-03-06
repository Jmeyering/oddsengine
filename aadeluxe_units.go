package oddsengine

func getDeluxeUnits() Units {
	return Units{
		Unit{
			Alias:      "lbr",
			Name:       "LONG RANGE BOMBER",
			Cost:       16,
			Attack:     7,
			Defend:     1,
			IsAircraft: true,
		},
		Unit{
			Alias:       "drt",
			Name:        "WWI DREADNOUGHT",
			Cost:        0,
			Attack:      4,
			Defend:      4,
			CanBombard:  true,
			CapitalShip: true,
		},
		Unit{
			Alias:            "aag",
			Name:             "ANTI-AIRCRAFT",
			Cost:             6,
			Attack:           0,
			Defend:           1,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "mif",
			Name:             "MOBILIZED INF",
			Cost:             6,
			Attack:           1,
			Defend:           1,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "inf",
			Name:             "INFANTRY",
			Cost:             2,
			Attack:           1,
			Defend:           1,
			CanTakeTerritory: true,
			PlusOneDefend: func(u map[string]int) int {
				var shots int

				numHif := numAllUnitsInFormation(u, "hif")
				numInf := numAllUnitsInFormation(u, "inf")

				hasCbf := numAllUnitsInFormation(u, "cbf") > 0
				hasMjb := numAllUnitsInFormation(u, "mjb") > 0
				hasMnb := numAllUnitsInFormation(u, "mnb") > 0

				totalBunkerCapacity := 0

				if hasCbf {
					totalBunkerCapacity += getDeluxeUnits().Find("cbf").Capacity
				}

				if hasMjb {
					totalBunkerCapacity += getDeluxeUnits().Find("mjb").Capacity
				}

				if hasMnb {
					totalBunkerCapacity += getDeluxeUnits().Find("mnb").Capacity
				}

				// Paired hif receive this bonus fist.
				if numHif >= totalBunkerCapacity {
					return shots
				}
				remainingShots := totalBunkerCapacity - numHif

				// Assume they will all be paired
				shots = numInf

				// If they can't all be paired, return the total number of
				// possible pairings
				if remainingShots < numInf {
					shots = remainingShots
				}

				return shots
			},
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var pairedArtilleryShotsAvailable int

				numInf := numAllUnitsInFormation(u, "inf")
				numLar := numAllUnitsInFormation(u, "lar")

				pairedArtilleryShotsAvailable = numLar

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
		},
		Unit{
			Alias:            "hif",
			Name:             "ELITE INFANTRY",
			Cost:             4,
			Attack:           2,
			Defend:           3,
			CanTakeTerritory: true,
			PlusOneDefend: func(u map[string]int) int {
				var shots int

				numHif := numAllUnitsInFormation(u, "hif")

				hasCbf := numAllUnitsInFormation(u, "cbf") > 0
				hasMjb := numAllUnitsInFormation(u, "mjb") > 0
				hasMnb := numAllUnitsInFormation(u, "mnb") > 0

				totalBunkerCapacity := 0

				if hasCbf {
					totalBunkerCapacity += getDeluxeUnits().Find("cbf").Capacity
				}

				if hasMjb {
					totalBunkerCapacity += getDeluxeUnits().Find("mjb").Capacity
				}

				if hasMnb {
					totalBunkerCapacity += getDeluxeUnits().Find("mnb").Capacity
				}

				shots = numHif
				if totalBunkerCapacity < numHif {
					shots = totalBunkerCapacity
				}

				return shots
			},
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var pairedArtilleryShotsAvailable int

				numInf := numAllUnitsInFormation(u, "hif")
				numLar := numAllUnitsInFormation(u, "lar")
				numHar := numAllUnitsInFormation(u, "har")

				pairedArtilleryShotsAvailable = numLar + numHar

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
		},
		Unit{
			Alias:            "lar",
			Name:             "LIGHT ARTILLERY",
			Cost:             4,
			Attack:           3,
			Defend:           2,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "har",
			Name:             "HEAVY ARTILLERY",
			Cost:             5,
			Attack:           4,
			Defend:           3,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "ltk",
			Name:             "LIGHT TANK",
			Cost:             4,
			Attack:           2,
			Defend:           2,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "htk",
			Name:             "HEAVY TANK",
			Cost:             6,
			Attack:           4,
			Defend:           4,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:      "flf",
			Name:       "FRONT LINE FIGHTER",
			Cost:       10,
			Attack:     4,
			Defend:     5,
			IsAircraft: true,
		},
		Unit{
			Alias:      "slf",
			Name:       "SECOND LINE FIGHTER",
			Cost:       0,
			Attack:     1,
			Defend:     1,
			IsAircraft: true,
		},
		Unit{
			Alias:      "tac",
			Name:       "TACTICAL BOMBER",
			Cost:       12,
			Attack:     4,
			Defend:     4,
			IsAircraft: true,
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var pairedTanksOrFlf int

				numTac := numAllUnitsInFormation(u, "tac")
				numLtk := numAllUnitsInFormation(u, "ltk")
				numHtk := numAllUnitsInFormation(u, "htk")
				numFlf := numAllUnitsInFormation(u, "flf")

				pairedTanksOrFlf = numLtk + numHtk + numFlf

				if pairedTanksOrFlf == 0 {
					return shots
				}

				// Assume they will all be paired
				shots = numTac

				// If they can't all be paired, return the total number of
				// possible pairings
				if pairedTanksOrFlf < numTac {
					shots = pairedTanksOrFlf
				}

				return shots
			},
		},
		Unit{
			Alias:      "sbr",
			Name:       "STRATEGIC BOMBER",
			Cost:       14,
			Attack:     6,
			Defend:     1,
			IsAircraft: true,
		},
		Unit{
			Alias:  "sbm",
			Name:   "SUBMARINE",
			Cost:   7,
			Attack: 3,
			Defend: 2,
			IsShip: true,
			IsSub:  true,
		},
		Unit{
			Alias:  "des",
			Name:   "DESTROYER",
			Cost:   9,
			Attack: 3,
			Defend: 3,
			IsShip: true,
		},
		Unit{
			Alias:      "csr",
			Name:       "CRUISER",
			Cost:       12,
			Attack:     5,
			Defend:     5,
			IsShip:     true,
			CanBombard: true,
		},
		Unit{
			Alias:       "acc",
			Name:        "AIRCRAFT CARRIER",
			Cost:        16,
			Attack:      0,
			Defend:      1,
			IsShip:      true,
			CapitalShip: true,
		},
		Unit{
			Alias:       "bts",
			Name:        "BATTLESHIP",
			Cost:        18,
			Attack:      6,
			Defend:      6,
			IsShip:      true,
			CapitalShip: true,
			CanBombard:  true,
		},
		Unit{
			Alias:    "cbf",
			Name:     "COASTAL BUNKER",
			Cost:     10,
			Attack:   3,
			Defend:   4,
			IsBunker: true,
			Capacity: 6,
		},
		Unit{
			Alias:    "mjb",
			Name:     "MAJOR BUNKER",
			Cost:     8,
			Attack:   2,
			Defend:   3,
			IsBunker: true,
			Capacity: 4,
		},
		Unit{
			Alias:    "mnb",
			Name:     "MINOR BUNKER",
			Cost:     10,
			Attack:   0,
			Defend:   0,
			IsBunker: true,
			Capacity: 3,
		},
	}
}
func get1940DeluxeUnits() Units {
	return Units{
		Unit{
			Alias:  "aag",
			Name:   "ANTI-AIRCRAFT GUN",
			Cost:   5,
			Attack: 0,
			Defend: 1,
			IsAAA:  true,
		},
		Unit{
			Alias:  "mif",
			Name:   "MECH Inf",
			Cost:   4,
			Attack: 1,
			Defend: 2,
			PlusOneRolls: func(u map[string]int) int {
				var shots int
				var remainingShots int

				var pairedArtilleryShotsAvailable int

				numInf := numAllUnitsInFormation(u, "inf")
				numMec := numAllUnitsInFormation(u, "mif")
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
			Alias:            "inf",
			Name:             "INFANTRY",
			Cost:             3,
			Attack:           1,
			Defend:           2,
			CanTakeTerritory: true,
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
		},
		Unit{
			Alias:            "art",
			Name:             "ARTILLERY",
			Cost:             4,
			Attack:           2,
			Defend:           2,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:            "tnk",
			Name:             "TANK",
			Cost:             6,
			Attack:           4,
			Defend:           4,
			CanTakeTerritory: true,
		},
		Unit{
			Alias:      "ftr",
			Name:       "FIGHTER",
			Cost:       10,
			Attack:     4,
			IsAircraft: true,
			Defend:     5,
		},
		Unit{
			Alias:      "tac",
			Name:       "TACTICAL BOMBER",
			Cost:       11,
			Attack:     4,
			Defend:     4,
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
			Alias:      "sbr",
			Name:       "STRATEGIC BOMBER",
			Cost:       12,
			Attack:     5,
			Defend:     1,
			IsAircraft: true,
		},
		Unit{
			Alias:  "sbm",
			Name:   "SUBMARINE",
			Cost:   6,
			Attack: 3,
			Defend: 1,
			IsShip: true,
			IsSub:  true,
		},
		Unit{
			Alias:  "des",
			Name:   "DESTROYER",
			Cost:   8,
			Attack: 2,
			Defend: 3,
			IsShip: true,
		},
		Unit{
			Alias:  "csr",
			Name:   "CRUISER",
			Cost:   12,
			Attack: 5,
			Defend: 5,
			IsShip: true,
		},
		Unit{
			Alias:       "acc",
			Name:        "AIRCRAFT CARRIER",
			Cost:        16,
			Attack:      1,
			Defend:      2,
			IsShip:      true,
			CapitalShip: true,
		},
		Unit{
			Alias:       "bts",
			Name:        "BATTLESHIP",
			Cost:        20,
			Attack:      6,
			Defend:      6,
			IsShip:      true,
			CanBombard:  true,
			CapitalShip: true,
		},
	}
}

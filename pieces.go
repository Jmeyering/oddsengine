package oddsengine

// Piece represents a specific piece within a game, identifing a lot of specific
// information related to the piece.
type Piece struct {
	Alias            string
	Name             string
	Cost             int
	Attack           int
	Defend           int
	IsShip           bool
	IsAircraft       bool
	CapitalShip      bool
	CanBombard       bool
	CanTakeTerritory bool
	PlusOneShots     func(map[string]int) int
}

// Pieces is a container for multiple Piece structs
type Pieces []Piece

// Swap implementing Sortable
func (p Pieces) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Len implementing Sortable
func (p Pieces) Len() int {
	return len(p)
}

// Create three different Sortable containers for our Pieces

// ByDefendingPower sorts the pieces by the higest Defend value of the unit
type ByDefendingPower struct{ Pieces }

// ByAttackingPower sorts the pieces by the higest Attack value of the unit
type ByAttackingPower struct{ Pieces }

// ByCost sorts the pieces by the lowest Cost value of the unit
type ByCost struct{ Pieces }

// Less implementing Sortable
func (p ByDefendingPower) Less(i, j int) bool {
	if p.Pieces[i].Alias == "aaa" {
		return false
	} else if p.Pieces[j].Alias == "aaa" {
		return true
	} else if p.Pieces[i].Defend == p.Pieces[j].Defend {
		return p.Pieces[i].Cost < p.Pieces[j].Cost
	}

	return p.Pieces[i].Defend < p.Pieces[j].Defend
}

// Less implementing Sortable
func (p ByAttackingPower) Less(i, j int) bool {
	if p.Pieces[i].Alias == "aaa" {
		return false
	} else if p.Pieces[j].Alias == "aaa" {
		return true
	} else if p.Pieces[i].Attack == p.Pieces[j].Attack {
		return p.Pieces[i].Cost < p.Pieces[j].Cost
	}

	return p.Pieces[i].Attack < p.Pieces[j].Attack
}

// Less implementing Sortable
func (p ByCost) Less(i, j int) bool {
	if p.Pieces[i].Alias == "aaa" {
		return false
	} else if p.Pieces[j].Alias == "aaa" {
		return true
	} else if p.Pieces[i].Cost == p.Pieces[j].Cost {
		return p.Pieces[i].Attack < p.Pieces[j].Attack
	}

	return p.Pieces[i].Cost < p.Pieces[j].Cost
}

// getPiecesForGame returns a Pieces type containing all the pieces that are
// valid for a particular game identified by the game string passed in
func getPiecesForGame(game string) (p Pieces) {
	switch game {
	case "1940":
		p = get1940Pieces()
	case "1941":
		p = get1941Pieces()
	case "1942":
		p = get1942Pieces()
	}
	return p
}

// Delete removes an item from a pieces type and returns the new type
func (p Pieces) Delete(alias string) Pieces {
	for i, piece := range p {
		if piece.Alias == alias {
			copy(p[i:], p[i+1:])
			p[len(p)-1] = Piece{}
			p = p[:len(p)-1]
		}
	}

	return p
}

// HasPiece tells us if a Pieces type has a piece identified by the alias
// provided
func (p Pieces) HasPiece(alias string) (has bool) {
	for _, piece := range p {
		if piece.Alias == alias {
			has = true
			break
		}
	}

	return has
}

// Find a specific piece, by alias, within the slice of Pieces
func (p Pieces) Find(alias string) *Piece {
	var targetPiece Piece
	for _, piece := range p {
		if piece.Alias == alias {
			targetPiece = piece
			break
		}
	}

	return &targetPiece
}

// get1940Pieces returns the pieces that are valid for the game Axis and Allies
// 1940
func get1940Pieces() Pieces {
	p := get1942Pieces()

	// The carrier in 1940 is now a capital ship
	p = p.Delete("car")

	p = append(p,
		// While not really a "piece" per se, we will treat it like one. It
		// should be deleted from any unit mapping once it is used
		Piece{
			Alias:  "kam",
			Name:   "Kamikaze",
			Cost:   0,
			Attack: 0,
			Defend: 2,
		},
		Piece{
			Alias:  "mec",
			Name:   "Mechanized Infantry",
			Cost:   4,
			Attack: 1,
			Defend: 2,
			// The mec defers all it's +1 shots to the inf, only taking whatever
			// shots are available
			PlusOneShots: func(u map[string]int) int {
				var shots int
				var remainingShots int

				numInf, hasInf := u["inf"]
				numMec := u["mec"]
				numArt, hasArt := u["art"]

				if !hasArt {
					return shots
				}

				// 2 mec 3 art
				if hasInf {
					remainingShots = numArt - numInf
				} else {
					remainingShots = numArt
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
		Piece{
			Alias:      "tac",
			Name:       "Tactical Bomber",
			Cost:       11,
			Attack:     3,
			Defend:     3,
			IsAircraft: true,
			PlusOneShots: func(u map[string]int) int {
				var shots int

				numTac := u["tac"]
				numFig, hasFig := u["fig"]
				numTan, hasTan := u["tan"]

				if !hasFig && !hasTan {
					return shots
				}

				totalNumBoosters := numFig + numTan

				// The number of shots are limited by the total number of tac
				// within the unit group
				shots = numTac
				if totalNumBoosters < numTac {
					shots = totalNumBoosters
				}

				return shots
			},
		},
		Piece{
			Alias:       "car",
			Name:        "Aircraft Carrier",
			Cost:        16,
			Attack:      0,
			Defend:      2,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: true,
		},
	)

	return p
}

// get1942Pieces returns the pieces that are valid for the game Axis and Allies
// 1942
func get1942Pieces() Pieces {
	p := get1941Pieces()
	// The 1942 Battleship gains the ability to bombard, it is also more
	// expensive
	p = p.Delete("bat")

	// The 1942 Battleship is more expensive
	p = p.Delete("car")
	p = append(p,
		Piece{
			Alias:  "aaa",
			Name:   "Anti-Aircraft Artillery",
			Cost:   5,
			Attack: 0,
			Defend: 1,
		},
		Piece{
			Alias:            "art",
			Name:             "Artillery",
			Cost:             4,
			Attack:           2,
			Defend:           2,
			CanTakeTerritory: true,
		},
		Piece{
			Alias:      "cru",
			Name:       "Cruiser",
			Cost:       12,
			Attack:     3,
			Defend:     3,
			IsShip:     true,
			CanBombard: true,
		},
		Piece{
			Alias:       "bat",
			Name:        "Battleship",
			Cost:        20,
			Attack:      4,
			Defend:      4,
			IsShip:      true,
			CanBombard:  true,
			CapitalShip: true,
		},
		Piece{
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

// get1941Pieces returns the pieces that are valid for the game Axis and Allies
// 1941
func get1941Pieces() Pieces {
	return Pieces{
		Piece{
			Alias:  "inf",
			Name:   "Infantry",
			Cost:   3,
			Attack: 1,
			Defend: 2,
			// Even though in 1941, there is no +1 for INF, having this here
			// does no harm because the conditions will never be met
			PlusOneShots: func(u map[string]int) int {
				var shots int

				numInf := u["inf"]
				numArt, hasArt := u["art"]

				if !hasArt {
					return shots
				}

				shots = numInf
				if numArt < numInf {
					shots = numArt
				}

				return shots
			},
			CanTakeTerritory: true,
		},
		Piece{
			Alias:            "tan",
			Name:             "Tank",
			Cost:             6,
			Attack:           3,
			Defend:           3,
			CanTakeTerritory: true,
		},
		Piece{
			Alias:      "fig",
			Name:       "Fighter",
			Cost:       10,
			Attack:     3,
			Defend:     4,
			IsAircraft: true,
		},
		Piece{
			Alias:      "bom",
			Name:       "Strategic Bomber",
			Cost:       12,
			Attack:     4,
			Defend:     1,
			IsAircraft: true,
		},
		Piece{
			Alias:  "sub",
			Name:   "Submarine",
			Cost:   6,
			IsShip: true,
			Attack: 2,
			Defend: 1,
		},
		Piece{
			Alias:  "des",
			Name:   "Destroyer",
			Cost:   8,
			Attack: 2,
			Defend: 2,
			IsShip: true,
		},
		Piece{
			Alias:       "car",
			Name:        "Aircraft Carrier",
			Cost:        12,
			Attack:      1,
			Defend:      2,
			IsShip:      true,
			CanBombard:  false,
			CapitalShip: false,
		},
		Piece{
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

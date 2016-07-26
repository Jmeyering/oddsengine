package oddsengine

import "sort"

var (
	landTroops   []string
	bombardShips []string
	capitalShips []string
	aircraft     []string
	surfaceShips []string
	ships        []string
	baseOol      []string
	noSubOol     []string
)

// setupOol creates all the piece slices that we will use within the engine.
func setupOol() {

	var hasAAA bool

	switch oolProfile {
	case "cost":
		sort.Sort(ByCost{activePieces})
	case "hitValue":
		sort.Sort(ByCost{activePieces})
	}

	// Range over all the active pieces for the specific game that is being
	// played and add them to their appropriate piece slices
	for _, p := range activePieces {
		// If the piece is an AAA we skip entirely. It needs to be added to the
		// baseOol last because of the special rules regarding when it can be
		// taken.
		if p.Alias == "aaa" {
			hasAAA = true
			continue
		}
		if p.CanTakeTerritory {
			landTroops = append(landTroops, p.Alias)
		}
		if p.CanBombard {
			bombardShips = append(bombardShips, p.Alias)
		}
		if p.CapitalShip {
			capitalShips = append(capitalShips, p.Alias)
		}
		if p.IsAircraft {
			aircraft = append(aircraft, p.Alias)
		}
		if p.IsShip && p.Alias != "sub" {
			surfaceShips = append(surfaceShips, p.Alias)
		}
		if p.IsShip {
			ships = append(ships, p.Alias)
		}
		if p.Alias != "sub" {
			noSubOol = append(noSubOol, p.Alias)
		}
		baseOol = append(baseOol, p.Alias)
	}

	// Every OOL that we create must add the "aaa" last. because AAA is a
	// special piece that must always be taken last.
	if hasAAA {
		baseOol = append(baseOol, "aaa")
	}

}

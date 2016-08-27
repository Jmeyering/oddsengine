package oddsengine

import (
	"sort"
	"strings"
)

var (
	landTroops     []string
	bombardShips   []string
	capitalShips   []string
	aircraft       []string
	subs           []string
	surfaceShips   []string
	ships          []string
	baseOol        []string
	noSubOol       []string
	multiRollUnits []string
)

func resetOol() {
	landTroops = []string{}
	bombardShips = []string{}
	capitalShips = []string{}
	aircraft = []string{}
	subs = []string{}
	surfaceShips = []string{}
	ships = []string{}
	baseOol = []string{}
	noSubOol = []string{}
	multiRollUnits = []string{}
}

// setupOol creates all the unit slices that we will use within the engine.
func setupOol() {

	resetOol()

	var hasAAA bool

	switch oolProfile {
	case "cost":
		sort.Sort(ByCost{activeUnits})
	case "hitValue":
		sort.Sort(ByCost{activeUnits})
	}

	// Range over all the active unit for the specific game that is being
	// played and add them to their appropriate unit slices
	for _, p := range activeUnits {
		// If the unit is an AAA we skip entirely. It needs to be added to the
		// baseOol last because of the special rules regarding when it can be
		// taken.
		if p.Alias == "aaa" || p.Alias == "raaa" {
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
		if p.IsShip && !p.IsSub {
			surfaceShips = append(surfaceShips, p.Alias)
		}
		if p.IsShip {
			ships = append(ships, p.Alias)
		}
		if p.IsSub {
			subs = append(subs, p.Alias)
		}
		if !p.IsSub {
			noSubOol = append(noSubOol, p.Alias)
		}
		if p.MultiRoll > 0 {
			multiRollUnits = append(multiRollUnits, p.Alias)
		}
		baseOol = append(baseOol, p.Alias)
	}

	// Every OOL that we create must add the "aaa" last. because AAA is a
	// special unit that must always be taken last.
	if hasAAA {
		baseOol = append(baseOol, "aaa", "raaa")
	}

}

// customizeOol takes the system's baseOol and customizes it for the particular
// units that have been passed in. The primary function in the real world is
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
		if strings.HasPrefix(alias, "+") && !sliceHas(ool, alias) {
			ool = append(ool, alias)
		}
	}

	// AAA Is always the last thing taken in any conflict
	if activeUnits.HasUnit("aaa") {
		ool = append(ool, "aaa")
	}
	if activeUnits.HasUnit("raaa") {
		ool = append(ool, "raaa")
	}

	return ool
}

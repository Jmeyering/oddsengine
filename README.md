# Axis and Allies Odds Calculator

[![Build Status](https://travis-ci.org/Jmeyering/oddsengine.svg?branch=master)](https://travis-ci.org/Jmeyering/oddsengine)

## Introduction

The odds engine is a tool to calculate the odds of Axis and Allies conflicts.
Currently supports the following games.

* Axis and Allies 1941
* Axis and Allies 1942 Second Edition
* Axis and Allies 1940 (Pacific, Europe and Global) Second Edition

There is a simple interface for generating a summary of a conflict.

```go

import (
    "fmt",
	"github.com/jmeyering/oddsengine",
)

func main() {
    attackers := map[string]int{"inf": 2, "art": 1, "fig": 2}
    defenders := map[string]int{"aaa": 1, "inf": 2, "tan": 1, "tac": 1}

    // Set the game that this conflict should run against. Default is "1940"
    oddsengine.SetGame("1940")

    // Set the number of iterations that the simulation will run the conflict
    // default is 1000
    oddsengine.SetIterations(10000)

    summary, err := oddsengine.GetSummary(attackers, defenders)

    if err != nil {
        panic(err)
    }

    fmt.Printf("%+v", summary)
    /*
    &{
        AverageRounds:2.73
        AttackerWinPercentage:59.91
        DefenderWinPercentage:35.01
        DrawPercentage:5.08
        AAAHitsAverage:0.34
        KamikazeHitsAverage:0
        AttackerAvIpcLoss:18.38
        DefenderAvIpcLoss:20.69
    }
    */
}

```

## Unit Formation Mapping

Unit formations are maps of units to number of units, `map[string]int` to be
precise. The map is formatted as `unitalias: numunits`. So
`map[string]int{"des":2, "bat":3}` is a unit formation of 2 destroyers and 3
battleships. All units are aliased by taking the first 3 letters of their name.

**Note:**

Strategic bombers are identified by `bom` rather than `str`.

## Unit Designations

There are a couple special designations that you can assign to a unit which
will affect how it functions and/or its order or loss position.

### Reserved

A reserved unit is designated by prefixing the unit alias with a "+". A
reserved unit will be moved to the end of the order of loss. Meaning it will
be taken last.

```go
attackers := map[string]int{"inf": 2, "+tan": 1, "tan": 2, "fig": 2}
```

The above unit formation includes 3 tanks, with one of them identified as being
"reserved". When taking losses, all non-reserved units will be taken before the
reserved tank. So the Order of loss above will be
`[]string{"inf", "tan", "fig", "+tan"}`

### Damaged


A damaged unit is designated by prefixing the unit alias with a "-". A damaged
unit designation is a way to represent a damaged capital ship.  In Axis and
Allies 1940, Capital ships retain all damage until they are repaired at a naval
base. When calculating odds for a battle involving damaged capital ships, it is
important the system is able to differentiate between non damaged and damaged
ships.

```go
defenders := map[string]int{"des": 3, "-bat": 1, "-car": 1, "bat": 2}
```

The above unit formation includes 3 battleships, of which, 1 is damaged. It
also contains 1 damaged carrier

## Order of loss

Order of loss is a complicated matter to tackle. There are multiple ways units
could be taken off the board, and accounting for real time situation assessment
to optimize an attack is rather difficult, and out of the scope of what I'm
attempting to accomplish.

For now, the engine uses a "cost" model of loss. Meaning that the more expensive
a unit is in the game, the higher likelihood it will be taken last. I've found
this to be generally a fine method of assigning loss. The only hiccups really
come when dealing with defending bombers and in some sea battles, where you
would really like to sacrifice some carriers and not take your cruisers or
aircraft.

Potentially I may work in an attacker and defender specific "value" model of
loss which will factor in both cost and hit value of the unit.

**Note:**

The current workaround to this problem is manually assigning
[reserved units](#reserved).

## Combined Arms

Combined arms are calculated appropriately for each game.

* Infantry & Mechanized Infantry +1 attack when paired with Artillery
* Tactical Bomber +1 attack when paired with a Tank or Fighter
* Aircraft + Destroyer can hit subs

## Special Combat

The engine appropriately calculates all types of special combat within the
supported games.

* AAA Defence
* Submarine Surprise Attack
* Kamikaze Strike
* Offshore Bombardment

### AAA Defence

AAA can simply be passed in as a defending unit and the engine will calculate
the number of shots appropriately. Pass in the number of **AAA Units** not the
number of AAA shots. The engine will appropriately calculate the number of shots
factoring in both the number of defending AAA Units and the number of attacking
aircraft. Casualties will be removed immediately without a chance to fire back.

```go
attackers := map[string]int{"fig":1, "tac":1, "bom":1}
defenders := map[string]int{"aaa":1}
```

Will fire 3 AAA shots at the attacking aircraft

### Submarine Surprise Attack

Submarines eligible to fire a surprise attack will do so, and any casualties will
be removed immediately without a chance to fire back

### Kamikaze Strike
Kamikaze strikes are passed in as a defending unit and the engine will calculate
the strike appropriately. Pass through the number of tokens used as the unit
number and the engine will calculate casualties. Casualties will be removed
immediately without a chance to fire back.

`defenders := map[string]int{"kam": 1, "des": 2, "sub": 2}`

### Offshore Bombardment

In order to run a simulation involving offshore bombardment, simply include
all your ground/air units and valid bombard-able ships to the attackers unit
formation. The simulation will appropriately calculate the bombard hits during
the first round of conflict.

`attackers := map[string]int{"inf": 3, "tan": 2, "tac": 2, "cru": 1, "bat": 2}`

Will fire two offshore bombardment shots from the battleship.

**Note:**

According to the rules, offshore bombardment is limited to the number of units
offloaded into the territory via transport. The engine assumes you have already
done this calculation and will not limit the number of bombardments that have
been passed in. See [Caveats](#caveats).

## Caveats

Very little time was spent worrying about error handling in cases where using
a little brain power works just fine.

For example within the game it is totally possible to have the following
conflict:

```go
attackers := map[string]int{"inf": 3}
defenders := map[string]int{"des": 4}

summary, err := oddsengine.GetSummary(attackers, defenders)
```

Of course this conflict is absurd. There is no circumstance where 3 attacking
infantry could attack 4 defending destroyers.

The engine, however, will return you a summary of this conflict and not error.
The onus is on the user to provide correct input in these cases.  For the
record, however, the destroyers win ~95% of the time.

package oddsengine

import "sort"

// RollValue simply defines the basic structure of a thing that can be rolled
// and assign hits to the opposing units. Given a hitValue (The number that
// counts as a "hit") and a num (The number of rolls at that number)
type RollValue struct{ hitValue, num int }

// RollMap is the type that can hold multiple RollValue types
type RollMap []RollValue

// Len return the length of the map
func (r RollMap) Len() int {
	return len(r)
}

// Less show which rollmap is less than than another
func (r RollMap) Less(i, j int) bool {
	return r[i].hitValue < r[j].hitValue
}

// Swap change the position of two elements in the map
func (r RollMap) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// AddRoll will add a specific RollValue to a RollMap based on a hitValue and
// num. returning the new RollMap with value added
func (r RollMap) AddRoll(hitValue, num int) (newMap RollMap) {
	if r.HasValue(hitValue) {
		i, v := r.Get(hitValue)
		v.num = v.num + num
		// Stupid way to delete from a slice while avoiding a memory leak.
		copy(r[i:], r[i+1:])
		r[len(r)-1] = RollValue{}

		r = r[:len(r)-1]

		if v.num > 0 {
			r = append(r, *v)
		}
		sort.Sort(r)

		return r
	}

	newMap = append(r, RollValue{hitValue, num})
	sort.Sort(newMap)

	return newMap
}

// Reduce will reduce the num value of one RollValue within a RollMap. If a
// RollMap, for example, Contains a RollValue of {hitValue: 3, num:2} and we
// Reduce(3, 1). The resulting RollMap will contain {hitValue: 3, num: 1}.
// having reduced the number of rolls at the hitvalue of 3. The new RollMap will
// be returned by this function
func (r RollMap) Reduce(hitValue, num int) (newMap RollMap) {
	return r.AddRoll(hitValue, num*-1)
}

// HasValue returns whether or not a RollMap has an entry for a RollValue with a
// hitValue matching the passed in num.
func (r RollMap) HasValue(num int) (has bool) {
	for _, a := range r {
		has = a.hitValue == num
		if has {
			break
		}
	}
	return has
}

// Get returns a the index of and the specific RollValue within the RollMap
// with a hitValue matching the passed in num
func (r RollMap) Get(num int) (index int, rv *RollValue) {
	for i, a := range r {
		if a.hitValue == num {
			rv = &a
			index = i
			break
		}
	}
	return index, rv
}

package collection

import (
	"github.com/willf/bitset"
)

// Storage for a single field takes:
// 4 bytes to store the value (float32)
// 1 byte to index the value's location (uint8)
// ~negligible/unknown size/overhead for bitset
// and total overhead for each unique field value set for the hashmap above
// effectively this means the weight of the item fields
// is 5 * len(fieldIndexes) + size of bitset
type itemFields struct {
	fieldBitset  bitset.BitSet
	fieldIndexes []uint8
	fieldValues  []float32
}

// k maps to a string via fieldMap
func (i *itemFields) getField(k int) float64 {
	u := uint8(k)
	ok := i.fieldBitset.Test(uint(k))
	if !ok {
		// Compatibility.
		return 0.
	}

	// Loop through the fieldIndexes, looking for
	// an index of k.
	var numFields int = len(i.fieldIndexes)
	for j := 0; j < numFields; j++ {
		// if fieldIndexes[j] == u, look at
		// fieldValues[j]
		if i.fieldIndexes[j] == u {
			// We store it as float32, but coerce it
			// back to float64 when retrieving.
			return float64(i.fieldValues[j])
		}
	}

	panic("should not be here")
	return 0.
}

// Returns whether or not it was updated or set new,
// where true is updated.
func (i *itemFields) setField(k int, val float64) bool {
	u := uint8(k)
	ok := i.fieldBitset.Test(uint(k))
	if ok {
		// Reset the field that was previously set.
		// Loop through the fieldIndexes, looking for
		// an index of k.
		var numFields int = len(i.fieldIndexes)
		for j := 0; j < numFields; j++ {
			// if fieldIndexes[j] == u, look at
			// fieldValues[j]
			if i.fieldIndexes[j] == u {
				var old float64 = float64(i.fieldValues[j])
				i.fieldValues[j] = float32(val)
				// Only return true if the value is new.
				return old != val
			}
		}

		panic("should not be here")
	} else {
		// Set the bitmap to 1 to signify we have this field.
		i.fieldBitset.Set(uint(k))

		// u = uint8(k)
		// fieldIndexes points from k to val
		i.fieldIndexes = append(i.fieldIndexes, u)
		i.fieldValues = append(i.fieldValues, float32(val))

		// This should be true if the field was updated.
		// Logically, this would always be true, but technically the
		// default field value is 0.
		return val != 0
	}
}

// Returns number of updated or new fields.
func (i *itemFields) setFields(ks []int, vals []float64) int {
	// Create an array for the new fields.
	var numFields int = len(ks)
	var numNewFields int
	var updatedFields int

	for j := 0; j < numFields; j++ {
		var k int = ks[j]
		if i.fieldBitset.Test(uint(k)) {
			continue
		}

		numNewFields += 1
	}

	// We act as those we're creating new fields every time,
	// but if numNewFields = 0, zero memory is allocated and the
	// remaining logical comparisons are significantly less
	// computationally intensive than the code around it. For that reason,
	// we execute the same loop to keep the code clean while
	// this could be separated into two segments, one for
	// new fields and one for existing fields.
	var newFields []uint8 = make([]uint8, numNewFields)
	var newValues []float32 = make([]float32, numNewFields)

	// the current running index inside the array
	var j int
	var offset int = len(i.fieldIndexes)
	for z := 0; z < numFields; z++ {
		var k int = ks[z]

		if i.fieldBitset.Test(uint(k)) {
			// Updating an already existing field.
			// Get the index at which we set.
			var target uint8 = uint8(k)
			for search := 0; search < offset; search++ {
				// if fieldIndexes[search] == k, look at
				// fieldValues[search]
				if i.fieldIndexes[search] == target {
					var old float64 = float64(i.fieldValues[search])
					i.fieldValues[search] = float32(vals[z])

					// Increment and break.
					if old != vals[z] {
						updatedFields += 1
					}

					break
				}
			}

		} else {
			// This is for setting new fields.
			// index at offset + j for the new slice
			newFields[j] = uint8(k)
			newValues[j] = float32(vals[z])

			i.fieldBitset.Set(uint(k))

			j += 1

			// Technically speaking, if the field is 0, we set
			// it, but it's not an update.
			if vals[z] != 0 {
				updatedFields += 1
			}
		}
	}

	if numNewFields > 0 {
		i.fieldIndexes = append(i.fieldIndexes, newFields...)
		i.fieldValues = append(i.fieldValues, newValues...)
	}

	return updatedFields
}

func (i *itemFields) deleteField(k int) {
	u := uint8(k)
	ok := i.fieldBitset.Test(uint(k))
	if !ok {
		return
	}

	i.fieldBitset.SetTo(uint(k), false)

	// Loop through the fieldIndexes, looking for
	// an index of k, which is the target delete field.
	var numFields int = len(i.fieldIndexes)
	// the index to delete
	var target int = numFields
	for j := 0; j < numFields; j++ {
		// if fieldIndexes[j] == u, look at
		// fieldValues[j]
		if i.fieldIndexes[j] == u {
			target = j
			break
		}
	}

	// replace the deletable with the last field
	i.fieldIndexes[numFields-1], i.fieldIndexes[target] = i.fieldIndexes[target], i.fieldIndexes[numFields-1]
	i.fieldIndexes = i.fieldIndexes[:numFields-1]
	i.fieldValues[numFields-1], i.fieldValues[target] = i.fieldValues[target], i.fieldValues[numFields-1]
	i.fieldValues = i.fieldValues[:numFields-1]
}

// Returns the number of bytes it takes to store
// the fields for this item in memory.
func (i *itemFields) weight() int {
	// 4 for float32, 1 for uint8
	return 5*len(i.fieldIndexes) + int(i.fieldBitset.BinaryStorageSize())
}

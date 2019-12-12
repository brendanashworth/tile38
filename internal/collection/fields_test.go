package collection

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"testing"

	"github.com/tidwall/geojson"
)

type fieldobject struct {
	fieldNames  []string
	fieldValues []float64
}

func BenchmarkItemFields(b *testing.B) {
	// Generate N objects, each with k fields of probability p (random),
	// and set values randomly. This is more of a memory test.
	var coll *Collection = New()

	var k int = 100
	var fillRatio float64 = 0.1

	// We have k fields, which will be filled with probability fillRatio.
	// They are just 0...k-1 in string form so we don't actually need to
	// map them.

	// Generate n objects, just points.
	json := `{"type":"Point","coordinates":[190,90]}`
	pt, err := geojson.Parse(json, nil)
	if err != nil {
		b.Fatal(err)
	}

	var objs []fieldobject = make([]fieldobject, b.N)

	for i := 0; i < b.N; i++ {
		// set fields
		var numFields int = int(math.Floor(float64(k) * fillRatio))

		var fieldNames []string = make([]string, numFields)
		var values []float64 = make([]float64, numFields)

		// set random fields at a random position.
		// so start at a random position inside k, then increment so long
		// as j is less than numFields. take the modulo to find insert position.
		var start int = int(math.Floor(rand.Float64() * float64(k)))
		for j := 0; j < numFields; j++ {
			fieldNames[j] = strconv.Itoa((start + j) % k)
			values[j] = 1.0
		}

		objs[i] = fieldobject{
			fieldNames:  fieldNames,
			fieldValues: values,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		obj := objs[i]

		coll.Set(strconv.Itoa(i), pt, obj.fieldNames, obj.fieldValues)
	}
	b.StopTimer()

	// Print out the memory usage of the first object.
	bytesUsed := coll.TotalWeight()
	fmt.Printf("%d bytes used, %f bytes per object, %f bytes per object per field\n",
		bytesUsed, float64(bytesUsed)/float64(b.N), float64(bytesUsed)/(float64(b.N*k)*fillRatio))
}

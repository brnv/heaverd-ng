package libperfomance

import (
	"fmt"
)


// Generate value that belongs segment [0;1]
func Hash(input string) float64 {
        data := []byte(input)
        hashsum := sha1.Sum(data)
        var ret uint32 = 0

		// length of hashsum is always eq 20
        for i:=0; i<20; i++ {
                ret = ret + uint32(hashsum[19-i]) * uint32(math.Pow(16, float64(i)))
        }
        ret = ret % 1000
        return float64(ret)/1000.0
}


// normalization by minimal argument
func min_norm(a,b int) float64 {
        var min int
        if a < b {
                min = a
        } else {
                min = b
        }
        return float64(min) / float64(b)
}




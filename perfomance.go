package main
import (
	"fmt"
	"crypto/sha1"
	"math"
)

type Resources struct {
	name string
	// params section
	cpu uint32
	cpu_capacity int
	
	disk uint64
	disk_capacity int

	ram uint32
	ram_capacity int

	zfs_arc_max uint64
	MEM_MIN uint64

	control_op_time uint32
	OP_TIME_THRESHOLD int
	SLOWNESS int

	uptime uint32
	UPTIME_PERIOD int

	// weight section
	cpu_weight float32
	disk_weight  float32
	ram_weight float32
	speed_factor float32
	uptime_factor float32
	// result
	length float32
}

// b-normalization on minimal value
func min_norm(a,b uint32) float32 {
	var min uint32
	if a < b {
		min = a
	}
	else {
		min = b
	}
	return float32(min) / float32(b)
}

// take string in input, generate hash, convert it to dec and 
// return % 1000 / 1000
func hash(input string) float32 {
	data := []byte(input)
	hashsum := sha1.Sum(data)
	var ret uint32 = 0
	for i:=0; i<len(hashsum); i++ {
		ret = ret + uint32(hashsum[len(hashsum)-i-1]) * uint32(math.Pow(16, float64(i)))
	}
	ret = ret % 1000
	
	return float32(ret)/1000.0
}


func main() {
	fmt.Println(Resources{name:"test"})
}




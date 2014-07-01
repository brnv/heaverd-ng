package main

import (
	"fmt"
	"crypto/sha1"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"math"
	"time"
	"runtime"
)

type Resources struct {
	hostname string
	

	// params section
	// cpu in % used by containers (lxc-cgroup)
	cpu uint32
	// total cpu capacity in % (e.g. if n cores available, it will be n*100%)
	cpu_capacity int
	
	disk uint64
	// total disk space
	disk_capacity int

	ram uint32
	// total memory available
	ram_capacity int


	// zfs arc cache value
	zfs_arc_max uint64

	// 1Gb
	// system reservation of CPU (defaul 100 (%) = 1 core for example)
	CPU_MIN uint64
	// system reservation of disk
	DISK_MIN uint64
	// system reservation of memory
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
}

// raw function for obtaining ram capacity
func get_ram_capacity() int {
	cmd := exec.Command("/usr/bin/free", "-b")
	out, _ := cmd.Output()
	result, _ := strconv.Atoi(strings.Fields(string(out))[7])
	return result

}


// raw function for obtaining disk capacity
func get_disk_capacity() int {
	cmd := exec.Command("/usr/bin/df", "-P", "/")
	out, _ := cmd.Output()
	result, _ := strconv.Atoi(strings.Fields(string(out))[8])
	fmt.Println(strings.Fields(string(out)))
	return result
}

func get_zfs_arc_max_value() uint64 {
	return 0
}

func (res *Resources) Setup() {
	// get local hostname
	res.hostname,_ = os.Hostname()
	// get cpu capacity: find out number of total available cores and 
	// multiplicate it on 100%
	res.cpu_capacity = runtime.NumCPU() * 100
	// setup cpu reservation (default 1 core * 100%)
	res.CPU_MIN = 100

	// determine total size of disk in Kb
	res.disk_capacity = get_disk_capacity()
	// setup disk reservation in Kb
	res.DISK_MIN = 1
	

	// determine ram capacity in bytes
	res.ram_capacity = get_ram_capacity()
	// setup ram reservation
	res.MEM_MIN = 1073741824 

	// determine zfs arc max value
	res.zfs_arc_max = get_zfs_arc_max_value()
}


func (res *Resources) Refresh() {
	res.cpu = 0
	res.disk = 0
	res.ram = 0

	res.uptime = 0
	res.control_op_time = 0

	fmt.Printf("Refreshed: %v\n", res)
}

func (res *Resources) GetLength() float32 {
	return res.speed_factor * res.uptime_factor * res.cpu_weight * res.disk_weight * res.ram_weight
}


// b-normalization on minimal value
func min_norm(a,b uint32) float32 {
	var min uint32
	if a < b {
		min = a
	} else {
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

var res Resources


func RefreshRoutine () {
	for {
		res.Refresh()
		time.Sleep(time.Second)
	}

}


func main() {
	//fmt.Println(Resources{name:"test"})
	res.Setup()
	
	go RefreshRoutine()
	for {
		time.Sleep(time.Minute)
		fmt.Println("Ping!")
	}
}




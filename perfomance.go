package main

import (
	"fmt"
	"crypto/sha1"
	"os"
	"os/exec"
	"io/ioutil"
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
	cpu int
	// total cpu capacity in % (e.g. if n cores available, it will be n*100%)
	cpu_capacity int
	
	disk int
	// total disk space
	disk_capacity int

	ram uint32
	// total memory available
	ram_capacity int


	// zfs arc cache value
	zfs_arc_max int

	// 1Gb
	// system reservation of CPU (defaul 100 (%) = 1 core for example)
	CPU_MIN int
	// system reservation of disk
	DISK_MIN int
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
	out, err := cmd.Output()

	// TODO: make code pretty
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine ram capacity (%v)\n", err)
		os.Exit(1)
	}

	result, err := strconv.Atoi(strings.Fields(string(out))[7])
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine ram capacity (%v)\n", err)
		os.Exit(1)
	}
	return result

}


// raw function for obtaining disk capacity
func get_disk_capacity() int {
	// TODO: configure mountpoint
	cmd := exec.Command("/usr/bin/df", "-P", "/")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine disk space size (%v)\n", err)
		os.Exit(1)
	}

	// TODO: make code pretty
	result, err := strconv.Atoi(strings.Fields(string(out))[8])
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine disk capacity (%v)\n", err)
		os.Exit(1)
	}

	return result
}

// obtain free space on disk
func get_disk_free() int {
	// TODO: configure mountpoint
	cmd := exec.Command("/usr/bin/df", "-P", "/")
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine disk space size (%v)\n", err)
		os.Exit(1)
	}

	// TODO: make code pretty
	result, err := strconv.Atoi(strings.Fields(string(out))[10])
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine disk capacity (%v)\n", err)
		os.Exit(1)
	}

	return result
}




func get_zfs_arc_max_value() int {
 	var num int
	fi, err := os.Open("/sys/module/zfs/parameters/zfs_arc_max")
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine zfs_arc_max value (%v)\n", err)
		os.Exit(1)
	}
	fmt.Fscanf(fi, "%d", &num)
	fi.Close()
	return num
}

func (res *Resources) Setup() {
	// TODO: make configurable params editable in configfiles
	// get local hostname
	hostname,err := os.Hostname()

	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine hostname (%v)\n", err)
		os.Exit(1)
	}
	res.hostname = hostname

	// get cpu capacity: find out number of total available cores and 
	// multiplicate it on 100%
	res.cpu_capacity = runtime.NumCPU() * 100
	// setup cpu reservation (default 1 core * 100%)
	res.CPU_MIN = 100

	// determine total size of disk in Kb
	res.disk_capacity = get_disk_capacity()
	// setup disk reservation in Kb
	res.DISK_MIN = 10 
	

	// determine ram capacity in bytes
	res.ram_capacity = get_ram_capacity()
	// setup ram reservation in bytes
	res.MEM_MIN = 1073741824 

	// determine zfs arc max value
	res.zfs_arc_max = get_zfs_arc_max_value()
}


func get_cpu_lxc_usage() int {
	data, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/lxc/cpuacct.stat")
	if err != nil {
		fmt.Printf("ERROR OCCURED: couldn't determine zfs_arc_max value (%v)\n", err)
		os.Exit(1)
	}
	user, _ := strconv.Atoi(strings.Fields(string(data))[1])
	system, _ := strconv.Atoi(strings.Fields(string(data))[3])
	return user+system
}


func (res *Resources) GetLength() float32 {
	return res.speed_factor * res.uptime_factor * res.cpu_weight * res.disk_weight * res.ram_weight
}


// b-normalization on minimal value
func min_norm(a,b int) float32 {
	var min int
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
	cpu_step1, cpu_step2 := 0, 0

	for {
		cpu_step1, cpu_step2 = cpu_step2, get_cpu_lxc_usage()
		// raw: get middle value for cpu usage
		res.cpu = (cpu_step2 - cpu_step1 + res.cpu) / 2
		res.cpu_weight = 1.0 - min_norm(res.cpu, res.cpu_capacity - res.CPU_MIN)

		res.disk = get_disk_free()
		res.disk_weight = 1.0 - min_norm(int(float32(res.disk_capacity) * 0.1), res.disk)
		

		
		res.ram = 0

		res.uptime = 0
		res.control_op_time = 0

		fmt.Printf("Refreshed: %v\n", res)
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




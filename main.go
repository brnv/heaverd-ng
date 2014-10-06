package main

import (
	"os"

	"github.com/brnv/heaverd-ng/tracker"
	"github.com/brnv/heaverd-ng/webapi"
	"github.com/docopt/docopt-go"
	"github.com/op/go-logging"
	"github.com/zazab/zhash"

	"bufio"
	"sync"
	"time"
)

var (
	configPath = "/etc/heaverd-ng/heaverd-ng.conf.toml"
	log        = logging.MustGetLogger("heaverd-ng")
	logfile    = os.Stderr
	format     = "%{time:15:04:05.000000} %{pid} %{level:.8s} %{message}"
	config     = zhash.NewHash()
)

var usage = `Heaverd-ng.
	Usage:
	heaverd-ng [-h|--help]
	heaverd-ng --config=<path>

	Options:
	-h|--help			Show this screen.
	--config=<path>		Configuration file [default: /etc/heaverd-ng/heaverd-ng.conf.toml].`

func main() {
	args, _ := docopt.Parse(usage, nil, true, "heaverd-ng", false)

	if args["--config"] != nil {
		configPath = string(args["--config"].(string))
	}

	loglevel := logging.INFO
	logging.SetBackend(logging.NewLogBackend(logfile, "", 0))
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(loglevel, log.Module)

	err := readConfig(configPath)
	if err != nil {
		log.Fatal(err.Error())
	}

	webPort, _ := config.GetString("web", "port")
	templatesDir, _ := config.GetString("templates", "dir")
	clusterPort, _ := config.GetString("cluster", "port")
	clusterPools, _ := config.GetStringSlice("cluster", "pools")
	etcdPort, _ := config.GetString("etcd", "port")

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		webapi.Run(map[string]interface{}{
			"webPort":      webPort,
			"templatesDir": templatesDir,
			"clusterPort":  clusterPort,
			"seed":         time.Now().UnixNano(),
		})
		wg.Done()
	}()
	go func() {
		tracker.Run(map[string]interface{}{
			"clusterPort":  clusterPort,
			"clusterPools": clusterPools,
			"etcdPort":     etcdPort,
		})
		wg.Done()
	}()
	wg.Wait()
}

func readConfig(path string) error {
	f, err := os.Open(path)
	if err == nil {
		config.ReadHash(bufio.NewReader(f))
		return nil
	} else {
		return err
	}
}

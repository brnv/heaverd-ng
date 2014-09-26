package main

import (
	"heaverd-ng/tracker"
	"heaverd-ng/webapi"
	"os"
	"sync"
	"time"

	"github.com/docopt/docopt.go"
	"github.com/op/go-logging"
)

var (
	configPath = "/etc/heaverd-ng/heaverd-ng.conf.toml"
	log        = logging.MustGetLogger("heaverd-ng")
	logfile    = os.Stderr
	format     = "%{time:15:04:05.000000} %{pid} %{level:.8s} %{message}"
)

var usage = `Heaverd-ng.
	Usage:
	heaverd-ng [-h|--help]
	heaverd-ng --config=<path>

	Options:
	-h|--help			Show this screen.
	--config=<path>		Configuration file [default: /etc/heaverd-ng/heaverd-ng.conf.toml].`

func main() {
	args, _ := docopt.Parse(usage, nil, true, "Heaverd-ng 1.0", false)

	if args["--config"] != nil {
		configPath = string(args["--config"].(string))
	}

	loglevel := logging.INFO
	logging.SetBackend(logging.NewLogBackend(logfile, "", 0))
	logging.SetFormatter(logging.MustStringFormatter(format))
	logging.SetLevel(loglevel, log.Module)

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() { webapi.Run(configPath, time.Now().UnixNano(), log); wg.Done() }()
	go func() { tracker.Run(configPath, log); wg.Done() }()
	wg.Wait()
}

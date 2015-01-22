package heaver

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	"github.com/brnv/heaverd-ng/liblxc"
)

var (
	createArgs       = []string{"heaver", "-CSn", "", "--format", "json"}
	netInterfaceArgs = []string{"--net", "br0"}
	controlArgs      = []string{"heaver", "", ""}
	listArgs         = []string{"heaver", "-Lm", "--format", "json"}
	startArgs        = []string{"heaver", "-Sn"}
	stopArgs         = []string{"heaver", "-Tn"}
	destroyArgs      = []string{"heaver", "-Dn"}
	pushArg          = []string{"heaver", "-Pn"}
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

type Image struct {
	Updated string `json:"updated"`
	Size    int64  `json:"size"`
	ZfsPath string `json:"zfs_path"`
}

func Create(containerName string,
	image []string,
	key string,
	ipPredefined string,
) (lxc.Container, error) {
	createArgs[2] = containerName

	args := createArgs

	for _, i := range image {
		args = append(args, "-i")
		args = append(args, i)
	}

	if key != "" {
		args = append(args, "--raw-key")
		args = append(args, key)
	}

	if ipPredefined != "" {
		args = append(args, "--raw-net")
		args = append(args, ipPredefined)
	} else {
		for _, n := range netInterfaceArgs {
			args = append(args, n)
		}
	}

	output, err := getHeaverOutput(args)

	if err != nil {
		return lxc.Container{Status: "error"}, err
	}

	messages := strings.Split(string(output), "\n")

	heaverOutputJson := struct {
		Data struct {
			Ips        []string
			Filesystem string
			Mountpoint string
		}
	}{}

	err = json.Unmarshal([]byte(messages[0]), &heaverOutputJson)

	if err != nil {
		return lxc.Container{Status: "error"}, err
	}

	ip := ""
	if len(heaverOutputJson.Data.Ips) != 0 {
		ip = heaverOutputJson.Data.Ips[0]
	}

	container := lxc.Container{
		Name:       containerName,
		Status:     "created",
		Image:      image,
		Filesystem: heaverOutputJson.Data.Filesystem,
		Mountpoint: heaverOutputJson.Data.Mountpoint,
		Ip:         ip,
	}

	return container, nil
}

func Push(containerName string, image string) error {
	args := append(pushArg, containerName)
	args = append(args, "-i")
	args = append(args, image)

	_, err := getHeaverOutput(args)

	return err
}

func Start(containerName string) error {
	args := append(startArgs, containerName)

	_, err := getHeaverOutput(args)

	return err
}

func Stop(containerName string) error {
	args := append(stopArgs, containerName)

	_, err := getHeaverOutput(args)

	return err
}

func Destroy(containerName string) error {
	args := append(destroyArgs, containerName)

	_, err := getHeaverOutput(args)

	return err
}

func ListContainers(host string) (map[string]lxc.Container, error) {
	output, err := getHeaverOutput(listArgs)
	if err != nil {
		return nil, err
	}

	heaverOutputJson := struct {
		Data map[string]struct {
			Active     bool
			Ips        map[string][]string
			Filesystem string
			Mountpoint string
		}
	}{}

	err = json.Unmarshal(output, &heaverOutputJson)

	list := map[string]lxc.Container{}

	for name, container := range heaverOutputJson.Data {
		status := StatusActive
		if container.Active == false {
			status = StatusInactive
		}

		list[name] = lxc.Container{
			Name:       name,
			Host:       host,
			Status:     status,
			Ips:        container.Ips,
			Filesystem: container.Filesystem,
			Mountpoint: container.Mountpoint,
		}
	}

	return list, nil
}

func ListImages() (map[string]Image, error) {
	cmd := exec.Command("heaver-img", "-Qj")
	raw, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	jsonResp := struct {
		Status string          `json:"status"`
		Data   json.RawMessage `json:"data"`
		Error  string          `json:"error"`
	}{}

	err = json.Unmarshal(raw, &jsonResp)
	if err != nil {
		return nil, err
	}

	if jsonResp.Error != "" {
		return nil, errors.New(jsonResp.Error)
	}

	list := make(map[string]Image)
	err = json.Unmarshal(jsonResp.Data, &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func getHeaverOutput(args []string) ([]byte, error) {
	cmd := &exec.Cmd{
		Path: "/usr/bin/heaver",
		Args: args,
	}

	var output bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &output
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil && stderr.String() != "" {
		return nil, errors.New(stderr.String())
	} else if err != nil {
		return nil, errors.New(output.String())
	}

	return output.Bytes(), nil
}

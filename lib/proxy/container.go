package proxy

import (
	"encoding/json"
	"fmt"
)

type Container interface {
	Build() bool
	Pull() bool
	Deploy() bool
	Run() bool
	Collect() bool
	Destroy() bool
	Status() string
}

type ContainerCommon struct {
	Distribution string
	Version      string
	Name         string
	BuildDir     string
	DeployDir    string
	RunCommand   string
}

func (cc ContainerCommon) Build() bool {
	fmt.Println("Basic build")
	return true
}

func ContainerNew(value string) (Container, bool) {
	var cc ContainerCommon
	var valid bool
	err := json.Unmarshal([]byte(value), &cc)
	if err != nil {
		valid = false
		return nil, valid
	} else {
		valid = true
	}

	switch cc.Distribution {
	case "docker":
		docker := DockerContainer{cc}
		return docker, valid
	case "oci":
		oci := OCIContainer{cc}
		return oci, valid
	case "rkt":
		rkt := RKTContainer{cc}
		return rkt, valid
	default:
		fmt.Println(cc.Distribution, "is not supported")
		valid = false
	}
	return nil, valid
}

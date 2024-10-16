package gadget

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const ECMFunctionTypeName = "ecm"

type ECMFunction struct {
	name     string
	path     string
	instance string

	g *Gadget
}

func (e *ECMFunction) Path() string {
	return e.path
}

func (e *ECMFunction) Name() string {
	return e.name
}

type ECMFunctionAttrs struct {
	HostAddr string
	DevAddr  string
}

func CreateECMFunction(g *Gadget, instance string) *ECMFunction {
	basePath := filepath.Join(g.Path(), g.Name(), FunctionsDir)
	name := fmt.Sprintf("%s.%s", ECMFunctionTypeName, instance)
	path := filepath.Join(basePath, name)

	function := &ECMFunction{
		name:     name,
		path:     basePath,
		instance: instance,

		g: g,
	}

	err := os.Mkdir(path, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	return function
}

func (h *ECMFunction) SetAttrs(attrs *ECMFunctionAttrs) {
	WriteString(h.path, h.name, "host_addr", attrs.HostAddr)
	WriteString(h.path, h.name, "dev_addr", attrs.DevAddr)
}

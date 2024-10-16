package gadget

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

const HidFunctionTypeName = "hid"

type HidFunction struct {
	name     string
	path     string
	instance string

	gadget *Gadget
}

func (h *HidFunction) Path() string {
	return h.path
}

func (h *HidFunction) Name() string {
	return h.name
}

type HidFunctionAttrs struct {
	Subclass     uint8
	Protocol     uint8
	ReportLength uint16
	ReportDesc   []byte
}

func CreateHidFunction(gadget *Gadget, instance string) (*HidFunction, error) {
	basePath := filepath.Join(gadget.Path(), gadget.Name(), FunctionsDir)
	name := fmt.Sprintf("%s.%s", HidFunctionTypeName, instance)
	path := filepath.Join(basePath, name)

	function := &HidFunction{
		name:     name,
		path:     basePath,
		instance: instance,

		gadget: gadget,
	}

	err := os.Mkdir(path, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("cannot create HID function: %w", err)
	}

	return function, nil
}

func (h *HidFunction) SetAttrs(attrs *HidFunctionAttrs) {
	WriteDec(h.path, h.name, "protocol", int(attrs.Protocol))
	WriteDec(h.path, h.name, "subclass", int(attrs.Subclass))
	WriteDec(h.path, h.name, "report_length", int(attrs.ReportLength))
	WriteBuf(h.path, h.name, "report_desc", attrs.ReportDesc)
}

func (h *HidFunction) GetReadWriter() (*bufio.ReadWriter, error) {
	file, err := os.OpenFile("/dev/hidg0", os.O_RDWR, 0666)
	if err != nil {
		return nil, err
	}

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(file)

	return bufio.NewReadWriter(reader, writer), nil
}

func (h *HidFunction) GetDev() (string, error) {
	path := filepath.Join(h.path, h.name, "dev")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("cannot get device: %w", err)
	}
	return string(data), nil
}

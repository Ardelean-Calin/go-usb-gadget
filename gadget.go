package gadget

import (
	"fmt"
	"os"
	"path/filepath"

	o "github.com/ardelean-calin/go-usb-gadget/option"
)

const BasePath = "/sys/kernel/config/usb_gadget"
const StringsDir = "strings"
const LangUsEng = 0x0409

type Gadget struct {
	path string
	name string
	udc  string

	enabled bool

	configs   []*Config
	functions []Function
	strings   []string
}

type GadgetAttrs struct {
	BcdUSB          o.Option[uint16]
	BDeviceClass    o.Option[uint8]
	BDeviceSubClass o.Option[uint8]
	BDeviceProtocol o.Option[uint8]
	BMaxPacketSize0 o.Option[uint8]
	IdVendor        o.Option[uint16]
	IdProduct       o.Option[uint16]
	BcdDevice       o.Option[uint16]
}

type GadgetStrs struct {
	SerialNumber string
	Manufacturer string
	Product      string
}

func CreateGadget(name string) (*Gadget, error) {
	path := filepath.Join(BasePath, name)

	gadget := &Gadget{
		path: BasePath,
		name: name,
	}

	err := os.Mkdir(path, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("cannot create gadget: %w", err)
	}

	return gadget, nil
}

func (g *Gadget) Enable(udc string) {
	WriteString(g.path, g.name, "UDC", udc)
	g.udc = udc
	g.enabled = true
}

func (g *Gadget) Disable() {
	WriteString(g.path, g.name, "UDC", "\n")
	g.udc = ""
	g.enabled = false
}

func (g *Gadget) SetAttrs(attrs *GadgetAttrs) {
	if attrs.BcdUSB.IsSome() {
		g.writeHex16("bcdUSB", attrs.BcdUSB.Value())
	}

	if attrs.BDeviceClass.IsSome() {
		g.writeHex8("bDeviceClass", attrs.BDeviceClass.Value())
	}

	if attrs.BDeviceSubClass.IsSome() {
		g.writeHex8("bDeviceSubClass", attrs.BDeviceSubClass.Value())
	}

	if attrs.BDeviceProtocol.IsSome() {
		g.writeHex8("bDeviceProtocol", attrs.BDeviceProtocol.Value())
	}

	if attrs.BMaxPacketSize0.IsSome() {
		g.writeHex8("bMaxPacketSize0", attrs.BMaxPacketSize0.Value())
	}

	if attrs.IdVendor.IsSome() {
		g.writeHex16("idVendor", attrs.IdVendor.Value())
	}

	if attrs.IdProduct.IsSome() {
		g.writeHex16("idProduct", attrs.IdProduct.Value())
	}

	if attrs.BcdDevice.IsSome() {
		g.writeHex16("bcdDevice", attrs.BcdDevice.Value())
	}
}

func (g *Gadget) SetStrs(strs *GadgetStrs, lang int) error {
	langStr := fmt.Sprintf("0x%x", lang)
	path := filepath.Join(g.path, g.name, StringsDir, langStr)

	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("cannot set strings: %w", err)
	}
	g.strings = append(g.strings, path)

	WriteString(path, "", "serialnumber", strs.SerialNumber)
	WriteString(path, "", "manufacturer", strs.Manufacturer)
	WriteString(path, "", "product", strs.Product)

	return nil
}

func (g *Gadget) writeHex16(file string, value uint16) {
	WriteHex16(g.path, g.name, file, value)
}

func (g *Gadget) writeHex8(file string, value uint8) {
	WriteHex8(g.path, g.name, file, value)
}

func (g *Gadget) writeBuf(file string, buf []byte) {
	WriteBuf(g.path, g.name, file, buf)
}

func (g *Gadget) Path() string {
	return g.path
}

func (g *Gadget) Name() string {
	return g.name
}

func (g *Gadget) IsEnabled() bool {
	return g.enabled
}

// CleanUp cleans up the USB gadget
func (g *Gadget) CleanUp() error {
	// We need to disable it first
	if g.IsEnabled() {
		g.Disable()
	}

	for _, c := range g.configs {
		// 1. Remove functions from configurations (aka the symlinks)
		for _, b := range c.bindings {
			linkPath := filepath.Join(b.config.path, b.name)
			fmt.Printf("Removing symlink: %s\n", linkPath)
			err := os.Remove(linkPath)
			if err != nil {
				return fmt.Errorf("cannot unlink %q: %w", linkPath, err)
			}
		}
		// 2. Remove strings directories in configurations
		for _, path := range c.strings {
			err := os.RemoveAll(path)
			if err != nil {
				return fmt.Errorf("cannot remove strings %q: %w", path, err)
			}
		}
		// 3. Remove the configurations
		err := os.RemoveAll(c.path)
		if err != nil {
			return fmt.Errorf("cannot remove configuration %q: %w", c.name, err)
		}
	}

	// 4. Remove the functions
	for _, f := range g.functions {
		err := os.RemoveAll(f.Path())
		if err != nil {
			return fmt.Errorf("cannot remove function %q: %w", f, err)
		}
	}

	// 5. Remove strings directories in the gadget
	for _, s := range g.strings {
		err := os.RemoveAll(s)
		if err != nil {
			return fmt.Errorf("cannot remove gadget string %q: %w", s, err)
		}
	}

	// 6. Remove the gadget
	err := os.RemoveAll(g.path)
	if err != nil {
		return fmt.Errorf("cannot remove gadget %q: %w", g.name, err)
	}

	fmt.Printf("Successfully removed gadget %q", g.name)

	return nil
}

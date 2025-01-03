package ipmi

import (
	"context"
	"fmt"
)

// 28.12 Set System Boot Options Command
type SetSystemBootOptionsRequest struct {
	// Thus, the BMC will automatically clear a 'boot flags valid bit' if
	// a system restart is not initiated by a Chassis Control command
	// within 60 seconds +/- 10% of the valid flag being set.
	//
	// The BMC will also clear the bit on any system resets or power-cycles that
	// are not triggered by a System Control command.
	//
	// This default behavior can be temporarily overridden using the 'BMC boot flag valid bit clearing' parameter.
	// [7] - 1b = mark parameter invalid / locked
	// 0b = mark parameter valid / unlocked
	MarkParameterInvalid bool
	// [6:0] - boot option parameter selector
	ParameterSelector BootOptionParameterSelector

	BootOptionParameter BootOptionParameter
}

// Table 28-14, Boot Option Parameters

type SetSystemBootOptionsResponse struct {
}

func (req *SetSystemBootOptionsRequest) Pack() []byte {
	parameterData := req.BootOptionParameter.Pack(req.ParameterSelector)

	out := make([]byte, 1+len(parameterData))

	b := uint8(req.ParameterSelector)
	if req.MarkParameterInvalid {
		b = setBit7(b)
	} else {
		b = clearBit7(b)
	}
	packUint8(b, out, 0)

	packBytes(parameterData, out, 1)

	return out
}

func (req *SetSystemBootOptionsRequest) Command() Command {
	return CommandSetSystemBootOptions
}

func (res *SetSystemBootOptionsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{
		0x80: "parameter not supported",
		0x81: "attempt to set the 'set in progress' value (in parameter #0) when not in the 'set complete' state. (This completion code provides a way to recognize that another party has already 'claimed' the parameters)",
		0x82: "attempt to write read-only parameter",
	}
}

func (res *SetSystemBootOptionsResponse) Unpack(msg []byte) error {
	return nil
}

func (res *SetSystemBootOptionsResponse) Format() string {
	return ""
}

// This command is used to set parameters that direct the system boot following a system power up or reset.
// The boot flags only apply for one system restart. It is the responsibility of the system BIOS
// to read these settings from the BMC and then clear the boot flags
func (c *Client) SetSystemBootOptions(ctx context.Context, request *SetSystemBootOptionsRequest) (response *SetSystemBootOptionsResponse, err error) {
	response = &SetSystemBootOptionsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

// SetBootDevice set the boot device for next boot.
// persist of false means it applies to next boot only.
// persist of true means this setting is persistent for all future boots.
func (c *Client) SetBootDevice(ctx context.Context, bootDeviceSelector BootDeviceSelector, bootType BIOSBootType, persist bool) error {
	req := &SetSystemBootOptionsRequest{
		MarkParameterInvalid: false,
		ParameterSelector:    BOPS_BootFlags,
		BootOptionParameter: BootOptionParameter{
			BootFlags: &BOP_BootFlags{
				BootFlagsValid:     true,
				Persist:            persist,
				BIOSBootType:       bootType,
				BootDeviceSelector: bootDeviceSelector,
			},
		},
	}
	if _, err := c.SetSystemBootOptions(ctx, req); err != nil {
		return fmt.Errorf("SetSystemBootOptions failed, err: %s", err)
	}
	return nil
}

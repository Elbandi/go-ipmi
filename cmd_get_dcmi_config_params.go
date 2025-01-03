package ipmi

import (
	"context"
	"fmt"
)

// [DCMI specification v1.5] 6.1.3 Get DCMI Configuration Parameters Command
type GetDCMIConfigParamsRequest struct {
	ParamSelector DCMIConfigParamSelector
	SetSelector   uint8 // use 00h for parameters that only have one set
}

type GetDCMIConfigParamsResponse struct {
	MajorVersion  uint8
	MinorVersion  uint8
	ParamRevision uint8
	ParamData     []byte
}

func (req *GetDCMIConfigParamsRequest) Pack() []byte {
	out := make([]byte, 3)

	packUint8(GroupExtensionDCMI, out, 0)
	packUint8(uint8(req.ParamSelector), out, 1)
	packUint8(req.SetSelector, out, 2)

	return out
}

func (req *GetDCMIConfigParamsRequest) Command() Command {
	return CommandGetDCMIConfigParams
}

func (res *GetDCMIConfigParamsResponse) CompletionCodes() map[uint8]string {
	return map[uint8]string{}
}

func (res *GetDCMIConfigParamsResponse) Unpack(msg []byte) error {
	if len(msg) < 5 {
		return ErrUnpackedDataTooShortWith(len(msg), 5)
	}

	if err := CheckDCMIGroupExenstionMatch(msg[0]); err != nil {
		return err
	}

	res.MajorVersion, _, _ = unpackUint8(msg, 1)
	res.MinorVersion, _, _ = unpackUint8(msg, 2)
	res.ParamRevision, _, _ = unpackUint8(msg, 3)
	res.ParamData, _, _ = unpackBytes(msg, 4, len(msg)-4)

	return nil
}

func (res *GetDCMIConfigParamsResponse) Format() string {
	return ""
}

func (c *Client) GetDCMIConfigParams(ctx context.Context, paramSelector DCMIConfigParamSelector, setSelector uint8) (response *GetDCMIConfigParamsResponse, err error) {
	request := &GetDCMIConfigParamsRequest{
		ParamSelector: paramSelector,
		SetSelector:   setSelector,
	}
	response = &GetDCMIConfigParamsResponse{}
	err = c.Exchange(ctx, request, response)
	return
}

func (c *Client) GetDCMIConfigParamsFor(ctx context.Context, param DCMIConfigParameter) error {
	paramSelector, setSelector := param.DCMIConfigParameter()

	request := &GetDCMIConfigParamsRequest{ParamSelector: paramSelector, SetSelector: setSelector}
	response := &GetDCMIConfigParamsResponse{}
	if err := c.Exchange(ctx, request, response); err != nil {
		return err
	}

	if err := param.Unpack(response.ParamData); err != nil {
		return fmt.Errorf("unpack param (%s[%d]) failed, err: %w", paramSelector.String(), paramSelector, err)
	}

	return nil
}

func (c *Client) GetDCMIConfigurations(ctx context.Context) (*DCMIConfig, error) {
	dcmiConfig := &DCMIConfig{}

	{
		param := DCMIConfigParam_ActivateDHCP{}
		if err := c.GetDCMIConfigParamsFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiConfig.ActivateDHCP = param
	}

	{
		param := DCMIConfigParam_DiscoveryConfiguration{}
		if err := c.GetDCMIConfigParamsFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiConfig.DiscoveryConfiguration = param
	}

	{
		param := DCMIConfigParam_DHCPTiming1{}
		if err := c.GetDCMIConfigParamsFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiConfig.DHCPTiming1 = param
	}

	{
		param := DCMIConfigParam_DHCPTiming2{}
		if err := c.GetDCMIConfigParamsFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiConfig.DHCPTiming2 = param
	}

	{
		param := DCMIConfigParam_DHCPTiming3{}
		if err := c.GetDCMIConfigParamsFor(ctx, &param); err != nil {
			return nil, err
		}
		dcmiConfig.DHCPTiming3 = param
	}

	return dcmiConfig, nil
}

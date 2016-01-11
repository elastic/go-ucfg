package ucfg

import "github.com/mitchellh/mapstructure"

// unpack config into 'to'. Default values will already be set in 'to'.
func (c *Config) Materialize(to interface{}) error {
	// meta := &mapstructure.Metadata{}
	config := &mapstructure.DecoderConfig{
		ErrorUnused:      false,
		ZeroFields:       false,
		WeaklyTypedInput: false,
		Result:           to,
		TagName:          "config",
		// Metadata:         meta,
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	err = decoder.Decode(c.fields)
	if err != nil {
		return err
	}

	// TODO: check meta data for unsued fields
	return nil
}

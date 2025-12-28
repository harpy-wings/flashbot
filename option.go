package flashbot

type Option func(*flashbot) error

func WithBuilders(builders []string) Option {
	return func(f *flashbot) error {
		f.builders = builders
		return nil
	}
}

func WithRelayURL(relayURL string) Option {
	return func(f *flashbot) error {
		f.relayURL = relayURL
		return nil
	}
}

func WithChainID(chainID uint64) Option {
	return func(f *flashbot) error {
		f.chainID = chainID
		return nil
	}
}

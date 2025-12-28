package flashbot

import (
	"context"
	"crypto/ecdsa"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// flashbot implements the IFlashbot interface.
type flashbot struct {
	logger   logrus.FieldLogger
	tracer   trace.Tracer
	relayURL string
	chainID  uint64

	builders []string
	pk       *ecdsa.PrivateKey
	ethC     *ethclient.Client
	client   *http.Client
}

var _ IFlashbot = (*flashbot)(nil)

func New(ctx context.Context, opts ...Option) (IFlashbot, error) {
	f := &flashbot{}
	err := f.setDefaults(ctx)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if err := opt(f); err != nil {
			return nil, err
		}
	}
	err = f.init(ctx)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *flashbot) setDefaults(ctx context.Context) error {
	var err error
	f.tracer = otel.GetTracerProvider().Tracer("flashbot")
	f.logger = logrus.StandardLogger()
	f.client = http.DefaultClient
	f.relayURL = MainnetRelayURL
	f.pk, err = crypto.GenerateKey()
	if err != nil {
		return err
	}
	return nil
}

func (f *flashbot) init(ctx context.Context) error {
	_, span := f.tracer.Start(ctx, "flashbot.init")
	defer span.End()
	return nil
}

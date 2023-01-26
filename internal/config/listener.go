package config

import (
	"reflect"

	"github.com/olegfomenko/solana-go"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type ListenConf struct {
	ProgramId      solana.PublicKey `fig:"program_id"`
	FromTx         solana.Signature `fig:"from_tx"`
	DisableCatchup bool             `fig:"disable_catchup"`
	Chain          string           `fig:"chain"`
}

func (c *config) ListenConf() ListenConf {
	return c.lconf.Do(func() interface{} {
		config := ListenConf{DisableCatchup: true}

		if err := figure.Out(&config).
			With(figure.BaseHooks, solHooks).
			From(kv.MustGetStringMap(c.getter, "listen")).
			Please(); err != nil {
			panic(err)
		}

		return config
	}).(ListenConf)
}

var solHooks = figure.Hooks{
	"solana.PublicKey": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		return reflect.ValueOf(solana.MustPublicKeyFromBase58(v)), nil
	},
	"solana.Signature": func(raw interface{}) (reflect.Value, error) {
		v, err := cast.ToStringE(raw)
		if err != nil {
			return reflect.Value{}, errors.Wrap(err, "expected string")
		}

		if v == "" {
			return reflect.ValueOf(solana.Signature{}), nil
		}

		return reflect.ValueOf(solana.MustSignatureFromBase58(v)), nil
	},
}

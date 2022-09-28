package config

import (
	"reflect"

	"github.com/olegfomenko/solana-go"
	"github.com/spf13/cast"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

type ListenConf struct {
	ProgramId      solana.PublicKey `fig:"program_id"`
	FromTx         solana.Signature `fig:"from_tx"`
	DisableCatchup bool             `fig:"disable_catchup"`
}

type BridgeListener interface {
	ListenConf() ListenConf
}

type listener struct {
	getter kv.Getter
	once   comfig.Once
}

func NewBridgeListener(getter kv.Getter) BridgeListener {
	return &listener{
		getter: getter,
	}
}

func (l *listener) ListenConf() ListenConf {
	return l.once.Do(func() interface{} {
		config := ListenConf{DisableCatchup: true}

		if err := figure.Out(&config).
			With(figure.BaseHooks, solHooks).
			From(kv.MustGetStringMap(l.getter, "listen")).
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

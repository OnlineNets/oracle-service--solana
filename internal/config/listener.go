package config

import (
	"github.com/olegfomenko/solana-go"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type ListenConf struct {
	ProgramId      solana.PublicKey
	FromTx         solana.Signature
	DisableCatchup bool
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
		var config struct {
			FromTx         string `fig:"from_tx"`
			DisableCatchup bool   `fig:"disable_catchup"`
			ProgramId      string `fig:"program_id"`
		}

		if err := figure.Out(&config).From(kv.MustGetStringMap(l.getter, "listen")).Please(); err != nil {
			panic(err)
		}

		listen := ListenConf{
			ProgramId:      solana.MustPublicKeyFromBase58(config.ProgramId),
			DisableCatchup: config.DisableCatchup,
		}

		if !config.DisableCatchup {
			listen.FromTx = solana.MustSignatureFromBase58(config.FromTx)
		}
		return listen
	}).(ListenConf)
}

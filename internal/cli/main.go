package cli

import (
	"context"

	"github.com/alecthomas/kingpin"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/logan/v3"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/config"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/service"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/catchup"
	"gitlab.com/rarify-protocol/sol-saver-svc/internal/solana/listener"
)

func Run(args []string) bool {
	log := logan.New()

	defer func() {
		if rvr := recover(); rvr != nil {
			log.WithRecover(rvr).Error("app panicked")
		}
	}()

	cfg := config.New(kv.MustFromEnv())
	log = cfg.Log()

	app := kingpin.New("sol-saver-svc", "")

	runCmd := app.Command("run", "run command")
	serviceCmd := runCmd.Command("service", "run service") // you can insert custom help

	migrateCmd := app.Command("migrate", "migrate command")
	migrateUpCmd := migrateCmd.Command("up", "migrate db up")
	migrateDownCmd := migrateCmd.Command("down", "migrate db down")

	cmd, err := app.Parse(args[1:])
	if err != nil {
		log.WithError(err).Error("failed to parse arguments")
		return false
	}

	switch cmd {
	case serviceCmd.FullCommand():
		go func() {
			err := catchup.NewService(cfg).Catchup(context.Background())
			if err != nil {
				panic(err)
			}
		}()

		go listener.NewService(cfg).Listen(context.Background())
		err = service.NewSaverService(cfg).Run()
	case migrateUpCmd.FullCommand():
		err = MigrateUp(cfg)
	case migrateDownCmd.FullCommand():
		err = MigrateDown(cfg)
	default:
		log.Errorf("unknown command %s", cmd)
		return false
	}
	if err != nil {
		log.WithError(err).Error("failed to exec cmd")
		return false
	}
	return true
}

package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	logging "github.com/ipfs/go-log/v2"
	"github.com/urfave/cli/v2"

	"github.com/filecoin-project/lotus/api"
	lcli "github.com/filecoin-project/lotus/cli"
)

var log = logging.Logger("main")

func init() {
	logging.SetLogLevel("main", "INFO")
}

func fullnodeServices(cctx *cli.Context) lcli.ServicesAPI {
	var svcs lcli.ServicesAPI
	var err error
	for {
		svcs, err = lcli.GetFullNodeServices(cctx)
		if err == nil {
			break
		}
		log.Info("waiting for fullnode to become available")
		time.Sleep(time.Second * 10)
	}
	return svcs
}

func gatewayAPI(cctx *cli.Context) api.Gateway {
	var api api.Gateway
	var err error
	for {
		api, _, err = lcli.GetGatewayAPI(cctx)
		if err == nil {
			break
		}
		log.Info("waiting for gateway to become available")
		time.Sleep(time.Second * 10)
	}
	return api
}

func main() {
	app := &cli.App{
		Name: "example",
		Action: func(cctx *cli.Context) error {
			svcs := fullnodeServices(cctx)
			api := svcs.FullNodeAPI()
			// api := gatewayAPI(cctx)
			ch, err := api.ChainNotify(cctx.Context)
			if err != nil {
				return err
			}
			cmu := sync.Mutex{}
			ctr := map[string]int64{
				"total_blocks":  0,
				"total_changes": 0,
				"base_fee":      0,
			}
			go func() {
				for hcs := range ch {
					for _, hc := range hcs {
						cmu.Lock()
						num_blocks := len(hc.Val.Blocks())
						log.Infow("head change", "type", hc.Type, "num_blocks", num_blocks)
						ctr["total_blocks"] += int64(num_blocks)
						ctr["total_changes"] += 1
						cmu.Unlock()
					}
				}
			}()
			go func() {
				for range time.NewTicker(time.Minute).C {
					fee, err := svcs.GetBaseFee(cctx.Context)
					if err != nil {
						log.Warn("error getting base fee: %w", err)
						continue
					}
					cmu.Lock()
					ctr["base_fee"] = fee.Int64()
					cmu.Unlock()
				}
			}()
			return http.ListenAndServe(":9999", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(ctr)
			}))
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Warnf("%w", err)
	}
}

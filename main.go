package main

import (
	"github.com/IAmRiteshKoushik/tentacloid/pkg"
	"github.com/spf13/viper"
)

func main() {
	var err error

	pkg.Rabbit, err = pkg.NewBroker(viper.GetString("message_broker_url"))
	if err != nil {
		return
	}
	pkg.Log.Info("[OK]: Message broker initialized successfully")
}

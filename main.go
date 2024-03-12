package main

import (
	"net/http"

	//"github.com/etherria/bitcoin-tx-builder/bitcoin"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/etherria/bitcoin-tx-builder/bitcoin"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/", func(c echo.Context) error {

		txBuild := bitcoin.NewTxBuild(1, &chaincfg.TestNet3Params)
		txBuild.SingleBuild()

		return c.String(http.StatusOK, "Hello, World!")
	})

	s := http.Server{
		Addr:    ":8081",
		Handler: e,
		//ReadTimeout: 30 * time.Second, // customize http.Server timeouts
	}
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		//log.Fatal(err)
		e.Logger.Fatal(err)
	}
}

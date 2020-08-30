package main

import (
	"context"
	"github.com/bwmarrin/snowflake"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"os/signal"
)

var idNode *snowflake.Node

const epoch = 1577836800398

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	snowflake.Epoch = epoch

	readConfig()
	environment := viper.GetString("environment")
	if environment != "release" {
		log.Logger = log.Level(zerolog.TraceLevel)
		log.Info().Msg("Environment isn't 'release'; using trace level")
	}

	var err error
	nodeID := viper.GetInt64("node")
	idNode, err = snowflake.NewNode(nodeID)
	if err != nil {
		log.Fatal().Err(err).Int64("node_id", nodeID).Msg("Couldn't create id generator node")
	}

	if dsn := viper.GetString("sentry.dsn"); dsn != "" {
		log.Debug().Str("dsn", dsn).Msg("Initialising Sentry")
		err := sentry.Init(sentry.ClientOptions{
			Dsn: dsn,
		})
		if err != nil {
			log.Error().Err(err).Str("dsn", dsn).Msg("Initialising Sentry")
		}
	}

	dbConnect()

	gin.SetMode(environment)
	r := setupRouter()

	addr := viper.GetString("http.address")
	if certFile := viper.GetString("security.cert_file"); certFile != "" {
		log.Info().Bool("tls", true).Str("addr", addr).Msg("Starting HTTP server with TLS")

		// Let's assume key_file is present.
		keyFile := viper.GetString("security.key_file")
		go func() {
			if err := r.RunTLS(addr, certFile, keyFile); err != nil {
				log.Fatal().Bool("tls", true).Str("addr", addr).Err(err).Msg("Failed to start HTTP server")
			}
		}()
	}

	log.Info().Bool("tls", false).Str("addr", addr).Msg("Starting HTTP server")
	go func() {
		if err := r.Run(addr); err != nil {
			log.Fatal().Err(err).Msg("Starting HTTP server")
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	<-sigs

	log.Info().Msg("Interrupt received, gracefully exiting")
	_ = pg.Close(context.Background())
	_ = rdb.Close()
	os.Exit(0)
}

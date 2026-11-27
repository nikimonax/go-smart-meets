package main

import (
	"log"

	"github.com/alecthomas/kong"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"github.com/nikimonax/go-smart-meets/internal/server"
	"github.com/nikimonax/go-smart-meets/internal/server/config"
)

var options struct {
	Config string `short:"c" default:"config.yaml" help:"config file path"`
}

var k = koanf.New(".")

func main() {
	kong.Parse(
		&options,
		kong.UsageOnError(),
	)

	if err := k.Load(file.Provider(options.Config), yaml.Parser()); err != nil {
		log.Fatalf("error loading config: %v\n", err)
	}

	var cfg config.Config

	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("error parsing config: %v\n", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("bad config: %v\n", err)
	}

	app := server.New(&cfg)

	if err := app.Err(); err != nil {
		log.Fatalf("failed init app: %v\n", err)
	}

	app.Run()
}

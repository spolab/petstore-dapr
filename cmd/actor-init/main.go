/*
Copyright 2022 Alessandro Santini

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"os"
	"time"

	"github.com/nats-io/jsm.go"
	"github.com/nats-io/nats.go"
	"github.com/rs/zerolog/log"
)

// actor-init is a simple init container that does the NATS configuration for the actors
// without having to recur to the JetStream operator (which does not really make sense for this small deployment)
func main() {
	log.Info().Msg("Actor server init container starting")
	natsAddr := os.Getenv("NATS_ADDR")
	nc, err := nats.Connect(natsAddr)
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to nats")
	}
	mgr, err := jsm.New(nc, jsm.WithTimeout(10*time.Second))
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to nats")
	}
	_, err = mgr.LoadOrNewStream("vet")
	if err != nil {
		log.Fatal().Err(err).Msg("connecting to nats")
	}
	log.Info().Msg("Actor server init container complete")
}

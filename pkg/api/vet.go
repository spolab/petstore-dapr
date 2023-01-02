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
package api

import (
	"io"
	"net/http"

	"github.com/dapr/go-sdk/client"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spolab/petstore/pkg/command"
)

func Register(dapr client.Client, broker string, topic string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx := r.Context()
		log.Debug().Str("id", id).Msg("begin register")
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			String(w, http.StatusBadRequest, err.Error())
			return
		}
		//
		// Invoke the actor method
		//
		log.Debug().Str("id", id).Str("payload", string(bytes)).Msg("executing actor method")
		raw, err := dapr.InvokeActor(ctx, &client.InvokeActorRequest{ActorType: "vet", ActorID: id, Method: "register", Data: bytes})
		if err != nil {
			String(w, http.StatusInternalServerError, err.Error())
			log.Error().Str("id", id).Err(err).Msg("invoking actor method")
			return
		}
		//
		// Parse the response
		//
		log.Debug().Str("id", id).Msg("parsing the response")
		var res command.ActorResponse
		err = JsonFromBytes(raw.Data, &res)
		if err != nil {
			String(w, http.StatusInternalServerError, err.Error())
			log.Error().Str("id", id).Err(err).Msg("parsing the response")
			return
		}
		//
		// If the command is OK, take the events sent and publish them
		//
		if res.Status == command.StatusOK {
			for _, event := range res.Events {
				if err := dapr.PublishEvent(ctx, broker, topic, event, client.PublishEventWithContentType("application/cloudevents+json")); err != nil {
					String(w, http.StatusInternalServerError, err.Error())
					return
				}
			}
		}
		JSON(w, http.StatusAccepted, &res)
		log.Debug().Str("id", id).Msg("END register")
	}
}

// Reads the events of interest and updates the read caches as required
func OnEvent(dapr client.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			NoContent(w, http.StatusInternalServerError)
			return
		}
		log.Info().Str("payload", string(bytes)).Msg("received a VetRegistered event")
	}
}

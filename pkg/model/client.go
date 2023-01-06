package model

import (
	"context"
	"fmt"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/dapr/go-sdk/actor"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spolab/petstore/pkg/command"
	"github.com/spolab/petstore/pkg/event"
	"github.com/spolab/petstore/pkg/framework"
)

type ClientActor struct {
	framework.BaseEventSourcedAggregate `json:"-"`
	validate                            *validator.Validate `json:"-"`
	Salutation                          string              `json:"salutation"`
	Surname                             string              `json:"surname"`
	Name                                string              `json:"name"`
	Phone                               string              `json:"phone"`
	Email                               string              `json:"email"`
}

func (actor *ClientActor) Type() string {
	return "client"
}

// register a new client
func (actor *ClientActor) Register(ctx context.Context, cmd *command.RegisterClientCommand) ([]*cloudevents.Event, error) {
	err := actor.Lifecycle.Execute(actor, func() error {
		// The actor already exists
		if actor.Version != 0 {
			return fmt.Errorf("client id '%s' already exists", actor.ID())
		}
		// Validate command
		if err := actor.validate.Struct(cmd); err != nil {
			return err
		}
		// Append the events
		cr := &event.ClientRegistered{
			Id:         actor.ID(),
			Salutation: cmd.Salutation,
			Name:       cmd.Name,
			Surname:    cmd.Surname,
			Phone:      cmd.Phone,
			Email:      cmd.Email,
		}
		actor.AppendEvent(event.CloudEvent(event.FromSource("client"), event.OfType(event.TypeClientRegisteredV1), event.WithDataAsJSON(cr)))
		return nil
	})
	if err != nil {
		log.Error().Str("id", actor.ID()).Err(err).Msg("executing command")
	}
	return actor.UncommittedEvents(), err
}

func (actor *ClientActor) Apply(ces ...*cloudevents.Event) error {
	log.Info().Str("id", actor.ID()).Msg("begin apply")
	for index, ce := range ces {
		switch ce.Type() {
		case event.TypeClientRegisteredV1:
			var ev event.ClientRegistered
			if err := ce.DataAs(&ev); err != nil {
				return err
			}
			actor.Email = ev.Email
			actor.Name = ev.Name
			actor.Phone = ev.Phone
			actor.Salutation = ev.Salutation
			actor.Surname = ev.Surname
			actor.Version = index + 1
		default:
			log.Warn().Str("id", actor.ID()).Str("type", ce.Type()).Msg("unknown event type")
		}
	}
	log.Info().Str("id", actor.ID()).Msg("end apply")
	return nil
}

// Tests the invariants of the client
func (actor *ClientActor) Check() error {
	return nil
}

// create new instances of an client actor
func ClientActorFactory() actor.Server {
	log.Info().Msg("begin clientactorfactory")
	result := &ClientActor{
		validate: validator.New(),
	}
	result.ClearEvents()
	result.Lifecycle = framework.EventSourcedCommandLifecycle{
		Repository: framework.EventSourcedActorRepository{},
	}
	log.Info().Msg("end clientactorfactory")
	return result
}

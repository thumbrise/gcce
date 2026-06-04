package observable

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/thumbrise/gcce/op/emit/pipeline"
	"github.com/thumbrise/op-universal-schema-go/schema"
)

type TermPass struct {
	key         string
	term        schema.Term
	reflectType reflect.Type
	logger      *slog.Logger
	dropMe      bool
}

func NewTermPass(key string, logger *slog.Logger, reflectType reflect.Type, term schema.Term) *TermPass {
	return &TermPass{key: key, logger: logger, reflectType: reflectType, term: term}
}

func (t *TermPass) ID() string {
	return t.term.ID
}

func (t *TermPass) Kind() *schema.Kind {
	return new(*t.term.Kind)
}

func (t *TermPass) Required() bool {
	return t.term.Required != nil && *t.term.Required
}

func (t *TermPass) ReflectType() reflect.Type {
	return t.reflectType
}

func (t *TermPass) MapTraits(passFunc pipeline.TermPassFunc) error {
	if passFunc == nil {
		panic("pass function is unexpected nil")
	}

	if t.term.Trait == nil {
		return nil
	}

	for i, traitTerm := range t.term.Trait {
		key := fmt.Sprintf("%s.trait.%d", t.key, i)
		logger := t.logger.With(slog.String("key", key))
		traitTermPass := NewTermPass(key, logger, nil, traitTerm)

		err := passFunc(i, traitTermPass)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *TermPass) Rename(value string, reason string) {
	t.logger.Debug("ID change",
		slog.String("reason", reason),
		slog.String("previous", t.term.ID),
		slog.String("next", value),
	)

	t.term.ID = value
}

func (t *TermPass) SetRequired(value bool, reason string) {
	t.logger.Debug("Required change",
		slog.String("reason", reason),
		slog.Any("previous", t.term.Required),
		slog.Bool("next", value),
	)

	t.term.Required = new(value)
}

func (t *TermPass) AppendTrait(value schema.Term, reason string) {
	t.logger.Debug("Trait append",
		slog.String("reason", reason),
		slog.Any("value", value),
	)

	t.term.Trait = append(t.term.Trait, value)
}

func (t *TermPass) Drop(reason string) {
	t.logger.Debug("Drop",
		slog.String("reason", reason),
		slog.Any("me", t.term),
	)

	t.dropMe = true
}

func (t *TermPass) DropMe() bool {
	return t.dropMe
}

func (t *TermPass) Traits() []schema.Term {
	return t.term.Trait
}

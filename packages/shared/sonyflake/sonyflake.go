package sonyflake

import (
	"errors"
	"time"

	"github.com/sony/sonyflake"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Generator interface {
	NextID() (uint64, error)
}

type Params struct {
	fx.In
	Logger *zap.Logger
}

type sonyflakeGenerator struct {
	sf *sonyflake.Sonyflake
}

// Module exports the ID generator provider
// Provides a distributed ID generator using Sonyflake
var Module = fx.Module("sonyflake",
	fx.Provide(NewGenerator),
)

func (g *sonyflakeGenerator) NextID() (uint64, error) {
	if g.sf == nil {
		return 0, errors.New("sonyflake generator not initialized")
	}
	return g.sf.NextID()
}

func NewGenerator(p Params) (Generator, error) {
	settings := sonyflake.Settings{
		StartTime: time.Date(2022, time.October, 10, 0, 0, 0, 0, time.UTC),
	}

	sf := sonyflake.NewSonyflake(settings)
	if sf == nil {
		return nil, errors.New("failed to initialize Sonyflake")
	}

	p.Logger.Info("sonyflake initialized successfully")

	return &sonyflakeGenerator{sf: sf}, nil
}

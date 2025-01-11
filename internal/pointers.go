package internal

import (
	"time"

	"github.com/google/uuid"
)

func String(str string) *string   { return &str }
func Int(i int) *int              { return &i }
func Int64(i int64) *int64        { return &i }
func Float64(f float64) *float64  { return &f }
func UInt(i uint) *uint           { return &i }
func Bool(b bool) *bool           { return &b }
func Time(t time.Time) *time.Time { return &t }
func UUID(u uuid.UUID) *uuid.UUID { return &u }

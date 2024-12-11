package api2

import "context"

type Handler func(ctx context.Context, req any) (any, any)

type Middleware func(ctx context.Context, req any, next Handler) (any, any)

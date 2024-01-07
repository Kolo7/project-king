package stringsvc

import (
	"context"

	"github.com/recolabs/microgen/examples/svc/entity"
)

// @microgen middleware, logging, tracing, http, recovering, main
type StringService interface {
	// @logs-ignore ans, err
	Uppercase(ctx context.Context, stringsMap map[string]string) (ans string, err error)
	Count(ctx context.Context, text string, symbol string) (count int, positions []int, err error)
	// @logs-len comments
	TestCase(ctx context.Context, comments []*entity.Comment) (tree map[string]int, err error)
}

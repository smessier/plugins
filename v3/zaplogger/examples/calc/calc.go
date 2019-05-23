package calcapi

import (
	"context"

	calc "goa.design/plugins/v3/v3/zaplogger/examples/calc/gen/calc"
	log "goa.design/plugins/v3/v3/zaplogger/examples/calc/gen/log"
)

// calc service example implementation.
// The example methods log the requests and return zero values.
type calcsrvc struct {
	logger *log.Logger
}

// NewCalc returns the calc service implementation.
func NewCalc(logger *log.Logger) calc.Service {
	return &calcsrvc{logger}
}

// Add implements add.
func (s *calcsrvc) Add(ctx context.Context, p *calc.AddPayload) (res int, err error) {
	s.logger.Info("calc.add")
	return
}

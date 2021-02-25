package embedded

import (
	"context"
	"fmt"

	"github.com/figment-networks/indexer-scheduler/structures"
)

type SchedulerContractor interface {
	ScrapeLatest(ctx context.Context, ldr structures.LatestDataRequest) (ldResp structures.LatestDataResponse, er error)
}

type LastDataInternalTransport struct {
	sc SchedulerContractor
}

func NewLastDataInternalTransport(sc SchedulerContractor) *LastDataInternalTransport {
	return &LastDataInternalTransport{
		sc: sc,
	}
}

func (ld *LastDataInternalTransport) GetLastData(ctx context.Context, ldReq structures.LatestDataRequest) (ldr structures.LatestDataResponse, err error) {
	ldr, err = ld.sc.ScrapeLatest(ctx, ldReq)
	if err != nil {
		return ldr, &structures.RunError{Contents: fmt.Errorf("error getting response from ScrapeLatest:  %w", err)}
	}

	return ldr, err
}

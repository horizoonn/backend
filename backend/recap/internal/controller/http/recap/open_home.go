package recap

import (
	"context"

	"github.com/avito-hackaton-team/avito-recap/backend/recap/generated/recapapi"
)

func (s *recapServer) OpenHome(ctx context.Context) (*recapapi.Redirect, error) {
	return &recapapi.Redirect{
		Location: s.recapService.OpenHome(ctx),
	}, nil
}

package sync

import (
	"context"
	"seborama/pcloud/tracker/db"

	"github.com/pkg/errors"
)

type tracker interface {
	FindPCloudVsLocalMutations(ctx context.Context) (db.FSMutations, error)
}

type sdkClient interface{}

type localClient interface{}

type Sync struct {
	tracker      tracker
	pCloudClient sdkClient
	localClient  localClient
}

func NewSync(pCloudClient sdkClient, localClient localClient) *Sync {
	return &Sync{
		pCloudClient: pCloudClient,
		localClient:  localClient,
	}
}

func (s *Sync) Right(ctx context.Context) error {
	mutations, err := s.tracker.FindPCloudVsLocalMutations(ctx)
	if err != nil {
		return nil
	}

	for _, m := range mutations {
		switch m.Type {
		case db.MutationTypeCreated:
		case db.MutationTypeDeleted:
		case db.MutationTypeModified:
		case db.MutationTypeMoved:
		default:
			return errors.Errorf("unknown mutation type '%s'", string(m.Type))
		}
	}

	return nil
}

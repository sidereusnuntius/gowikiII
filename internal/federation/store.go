package federation

import (
	"context"

	"github.com/sidereusnuntius/gowiki/internal/model"
)

type Store interface {
	GetHost(ctx context.Context, hostname string) (model.Host, error)
	SaveHost(ctx context.Context, host *model.Host) error
	UpdateHostStatus(ctx context.Context, id int64, status model.HostStatus) error
}

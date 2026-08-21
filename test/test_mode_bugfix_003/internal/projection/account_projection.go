package projection

import (
	"context"
	"event-sourcing-service-test/internal/domain"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Accounts struct{ DB *pgxpool.Pool }

func (a Accounts) Apply(c context.Context, e domain.Event) error {
	if e.Type == domain.AccountOpened {
		_, err := a.DB.Exec(c, `insert into account_views(tenant_id,account_id,currency,balance,credited,debited,version,last_event_id) values($1,$2,$3,0,0,0,$4,$5) on conflict(tenant_id,account_id) do update set version=excluded.version,last_event_id=excluded.last_event_id where account_views.version < excluded.version`, e.TenantID, e.AggregateID, e.Payload["currency"], e.Version, e.ID)
		return err
	}
	n := domainInt(e.Payload["amount"])
	if e.Type == domain.MoneyCredited {
		result, err := a.DB.Exec(c, `update account_views set balance=balance+$1,credited=credited+$1,version=$2,last_event_id=$3 where tenant_id=$4 and account_id=$5 and version=$2-1`, n, e.Version, e.ID, e.TenantID, e.AggregateID)
		if err == nil && result.RowsAffected() == 0 {
			return a.duplicate(c, e)
		}
		return err
	}
	if e.Type != domain.MoneyDebited {
		return fmt.Errorf("unknown account event %q", e.Type)
	}
	result, err := a.DB.Exec(c, `update account_views set balance=balance-$1,debited=debited+$1,version=$2,last_event_id=$3 where tenant_id=$4 and account_id=$5 and version=$2-1`, n, e.Version, e.ID, e.TenantID, e.AggregateID)
	if err == nil && result.RowsAffected() == 0 {
		return a.duplicate(c, e)
	}
	return err
}
func (a Accounts) duplicate(c context.Context, e domain.Event) error {
	var version int64
	err := a.DB.QueryRow(c, `select version from account_views where tenant_id=$1 and account_id=$2`, e.TenantID, e.AggregateID).Scan(&version)
	if err != nil {
		return fmt.Errorf("account projection missing: %w", err)
	}
	if version >= e.Version {
		return nil
	}
	return fmt.Errorf("account projection out of order: have %d want %d", version, e.Version)
}
func domainInt(v any) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case int64:
		return x
	}
	return 0
}

package projection

import (
	"context"
	"encoding/json"
	"event-sourcing-service/internal/domain"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Orders struct{ DB *pgxpool.Pool }

func (o Orders) Apply(c context.Context, e domain.Event) error {
	status := map[string]string{domain.OrderCreated: "pending", domain.OrderPaid: "paid", domain.OrderCancelled: "cancelled", domain.OrderExpired: "expired"}[e.Type]
	if status == "" {
		return fmt.Errorf("unknown order event %q", e.Type)
	}
	b, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}
	if e.Type == domain.OrderCreated {
		_, err = o.DB.Exec(c, `insert into order_views(tenant_id,order_id,buyer_id,currency,amount,status,version,last_event_id) values($1,$2,coalesce(($3::jsonb->>'buyer_id'),''),coalesce(($3::jsonb->>'currency'),''),coalesce(($3::jsonb->>'amount')::bigint,0),$4,$5,$6) on conflict(tenant_id,order_id) do update set buyer_id=excluded.buyer_id,currency=excluded.currency,amount=excluded.amount,status=excluded.status,version=excluded.version,last_event_id=excluded.last_event_id where order_views.version < excluded.version`, e.TenantID, e.AggregateID, b, status, e.Version, e.ID)
		return err
	}
	result, err := o.DB.Exec(c, `update order_views set status=$1,version=$2,last_event_id=$3 where tenant_id=$4 and order_id=$5 and version=$2-1`, status, e.Version, e.ID, e.TenantID, e.AggregateID)
	if err == nil && result.RowsAffected() == 0 {
		var version int64
		queryErr := o.DB.QueryRow(c, `select version from order_views where tenant_id=$1 and order_id=$2`, e.TenantID, e.AggregateID).Scan(&version)
		if queryErr == nil && version >= e.Version {
			return nil
		}
		return fmt.Errorf("order projection missing or out of order for version %d", e.Version)
	}
	return err
}

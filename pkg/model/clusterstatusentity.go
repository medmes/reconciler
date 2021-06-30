package model

import (
	"fmt"
	"time"

	"github.com/kyma-incubator/reconciler/pkg/db"
)

const tblStatuses string = "inventory_cluster_config_statuses"

type ClusterStatusEntity struct {
	ID            int64         `db:"readOnly"`
	ConfigVersion int64         `db:"notNull"`
	Status        ClusterStatus `db:"notNull"`
	Created       time.Time     `db:"readOnly"`
}

func (c *ClusterStatusEntity) String() string {
	return fmt.Sprintf("ClusterStatusEntity [ConfigVersion=%d,Status=%s]",
		c.ConfigVersion, c.Status)
}

func (c *ClusterStatusEntity) New() db.DatabaseEntity {
	return &ClusterStatusEntity{}
}

func (c *ClusterStatusEntity) Marshaller() *db.EntityMarshaller {
	marshaller := db.NewEntityMarshaller(&c)
	marshaller.AddUnmarshaller("Status", func(value interface{}) (interface{}, error) {
		return NewClusterStatus(value.(string))
	})
	marshaller.AddUnmarshaller("Created", convertTimestampToTime)
	return marshaller
}

func (c *ClusterStatusEntity) Table() string {
	return tblStatuses
}

func (c *ClusterStatusEntity) Equal(other db.DatabaseEntity) bool {
	if other == nil {
		return false
	}
	otherClProp, ok := other.(*ClusterStatusEntity)
	if ok {
		return c.ConfigVersion == otherClProp.ConfigVersion && c.Status == otherClProp.Status
	}
	return false
}
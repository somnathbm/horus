package target

type TargetType string

const (
	HTTPTargetType        TargetType = "http"
	MySQLTargetType       TargetType = "mysql"
	PostgresSQLTargetType TargetType = "postgres"
	RedisTargetType       TargetType = "redis"
	KafkaTargetType       TargetType = "kafka"
)

type TargetSpec interface {
	isTargetSpec()
}

type Target struct {
	ID   string     `json:"id"`
	Name string     `json:"name"`
	Type TargetType `json:"type"`
	Spec TargetSpec `json:"spec"`
}

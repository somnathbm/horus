package target

type PostgresTargetSpec struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Port     int    `json:"port"`
}

func (postgres PostgresTargetSpec) isTargetSpec() {

}

package target

type RedisTargetSpec struct {
	URL  string `json:"url"`
	Port int    `json:"port"`
	TTL  int    `json:"ttl"`
}

func (redis RedisTargetSpec) isTargetSpec() {

}

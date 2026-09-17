package target

type HTTPTargetSpec struct {
	URL  string `json:"url"`
	Port int    `json:"port"`
}

func (http HTTPTargetSpec) isTargetSpec() {

}

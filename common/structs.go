package wysteria_common

type Collection struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
}

type Item struct {
	Parent   string            `json:"Parent"`
	Id       string            `json:"Id"`
	ItemType string            `json:"ItemType"`
	Variant  string            `json:"Variant"`
	Facets   map[string]string `json:"Facets"`
}

type Version struct {
	Parent    string            `json:"Parent"`
	Id        string            `json:"Id"`
	Number    int32               `json:"Number"`
	Facets    map[string]string `json:"Facets"`
}

type Resource struct {
	Parent       string `json:"Parent"`
	Name         string `json:"Name"`
	ResourceType string `json:"ResourceType"`
	Id           string `json:"Id"`
	Location     string `json:"Location"`
}

type Link struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
	Src  string `json:"Src"`
	Dst  string `json:"Dst"`
}

// Generic way to describe what we're searching for
type QueryDesc struct {
	Parent        string
	Id            string
	VersionNumber int32
	ItemType      string
	Variant       string
	Facets        map[string]string
	Name          string
	ResourceType  string
	Location      string
	LinkSrc       string
	LinkDst       string
}

type Marshalable interface {
	MarshalJSON() ([]byte, error)
}
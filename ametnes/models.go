package ametnes

type Resource struct {
	Id          int    `json:"id,omitempty"`
	Project     int    `json:"project,omitempty"`
	Account     int    `json:"account,omitempty"`
	Kind        string `json:"kind,omitempty"`
	Location    string `json:"location,omitempty"`
	Network     int    `json:"network,omitempty"`
	Name        string `json:"name,omitempty"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
	Product     int    `json:"product,omitempty"`
	Spec        Spec   `json:"spec,omitempty"`
}

type Spec struct {
	Components map[string]interface{} `json:"components,omitempty"`
	Nodes      int                    `json:"nodes,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Networks   []Networks             `json:"networks,omitempty"`
}
type Resources struct {
	Count int        `json:"count,omitempty"`
	Items []Resource `json:"results,omitempty"`
}

type Networks struct {
	Id int `json:"id,omitempty"`
}

type Project struct {
	Id          int    `json:"id,omitempty"`
	Account     int    `json:"account,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type Projects struct {
	Count int       `json:"count,omitempty"`
	Items []Project `json:"results,omitempty"`
}

type Product struct {
	Id   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type Capacity struct {
	Cpu     int
	Memory  int
	Storage int
}

func ProductFilter(ss []Product, test func(Product) bool) (ret []Product) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

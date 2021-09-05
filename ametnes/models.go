package ametnes

type Resource struct {
	Id          int                    `json:"id,omitempty"`
	Project     int                    `json:"project,omitempty"`
	Account     int                    `json:"account,omitempty"`
	Kind        string                 `json:"kind,omitempty"`
	Location    string                 `json:"location,omitempty"`
	Network     int                    `json:"network,omitempty"`
	Name        string                 `json:"name,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Description string                 `json:"description,omitempty"`
	Spec        map[string]interface{} `json:"spec,omitempty"`
}

type Resources struct {
	Count int        `json:"count,omitempty"`
	Items []Resource `json:"results,omitempty"`
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

func ProductFilter(ss []Product, test func(Product) bool) (ret []Product) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}

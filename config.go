package userconfig

type Configuration struct {
	Domains      DomainConfig      `json:"domains"`
	Certificates map[string]string `json:"certificates"`
	MetaData     map[string]string `json:"meta_data"`
}

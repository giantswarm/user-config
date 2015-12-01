package userconfig

type MetaDataConfig struct {
	Stack string `json:"stack"`
}

type ServiceConfig struct {
	MetaData MetaDataConfig `json:"meta_data"`
}

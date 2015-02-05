package userconfig

type MetaDataConfig struct {
	Stack string `json:"stack"`
}

type AppConfig struct {
	MetaData MetaDataConfig `json:"meta_data"`
}

package data

type DataFile struct {
	Data []DataBlock `yaml:"data"`
}

type DataBlock struct {
	ID      string `yaml:"id"`
	User    string `yaml:"user"`
	Content string `yaml:"content"`
}

package conf

import (
	"github.com/omeid/uconfig"
	"os"
)

type Config struct {
	NumAliens   int    `default:"5" usage:"enter the number of aliens" env:"NUM_ALIENS" flag:"num-aliens"`
	CityMapFile string `default:"input.txt" usage:"enter the file path of a city map" env:"CITY_MAP_FILE_PATH" flag:"city-map-file-path"`
}

func GetConfig() *Config {
	conf := &Config{}
	c, err := uconfig.Classic(conf, nil)
	if err != nil {
		c.Usage()
		os.Exit(1)
	}
	return conf
}

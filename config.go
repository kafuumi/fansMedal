package fans

import (
	"github.com/gf-third/yaml"
	"io"
)

type Config struct {
	List      []int64 `yaml:"list"` //过滤名单
	Type      bool    `yaml:"type"` //为true时，list为白名单，否则为黑名单
	AccessKey string  `yaml:"accessKey"`
	LikeCD    int     `yaml:"likeCD"` //点赞cd，默认3秒
	ChatCD    int     `yaml:"chatCD"` //发弹幕cd，默认5秒
	IsWear    bool    `yaml:"isWear"`
}

func ReadConfig(r io.Reader) (Config, error) {
	var cfg Config
	cfg.LikeCD = 3
	cfg.ChatCD = 3
	cfg.IsWear = true
	decoder := yaml.NewDecoder(r)
	err := decoder.Decode(&cfg)
	return cfg, err
}

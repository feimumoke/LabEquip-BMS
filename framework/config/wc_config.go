package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"

	"gopkg.in/yaml.v2"
)

type WeChatingConfig struct {
	DataSource      map[string]*ScormClusterConfig `yaml:"datasource"`
	KafkaConf       *KafKaConf                     `yaml:"kafka"`
	ESCfg           *ESClusterConfig               `yaml:"es"`
	DistributeCache map[string]*CacheItem          `yaml:"cache"`
	Server          *ServerConfig                  `yaml:"server"`
	JwtConfigs      []*JwtConfig                   `yaml:"jwtconfig"`
	Log             *LogConfig                     `yaml:"log"`
}

type ServerConfig struct {
	Addr map[string]string `yaml:"addr"`
}

type JwtConfig struct {
	Account   string `yaml:"account"`
	AppKey    string `yaml:"ak"`
	SecretKey string `yaml:"sk"`
}

type LogConfig struct {
	Dir           string `yaml:"dir"`
	Level         string `yaml:"level"`
	EnableConsole bool   `yaml:"enableConsole"`
	GormLogLevel  string `yaml:"gormLogLevel"`
}

var WCConfig = &WeChatingConfig{}

type CacheItem struct {
	Type    string `yaml:"type"`
	Address string `yaml:"address"`
}

type ScormClusterConfig struct {
	Driver            string              `yaml:"driver"`
	AutoReport        bool                `yaml:"autoReport"`
	BlockGlobalUpdate bool                `yaml:"blockGlobalUpdate"`
	LogDisabled       bool                `yaml:"logDisabled"`
	MaxIdleConns      int                 `yaml:"maxIdleConns"`
	MaxOpenConns      int                 `yaml:"maxOpenConns"`
	ConnMaxLifetime   int                 `yaml:"connMaxLifetime"`
	Groups            []*ScormGroupConfig `yaml:"groups"`
	ShadowGroups      []*ScormGroupConfig `yaml:"shadowGroups"`
}

type ScormGroupConfig struct {
	MasterDsn   string   `yaml:"masterDsn"`
	ReplicasDsn []string `yaml:"replicasDsn"`
}

type ConsumerConfItem struct {
	FromSystem    string   `yaml:"FromSystem"`
	Brokers       []string `yaml:"Brokers"`
	ConsumerGroup string   `yaml:"ConsumerGroup"`
	GoroutineSize int      `yaml:"GoroutineSize"`
}

type ProducerConfItem struct {
	ToSystem string                   `yaml:"ToSystem"`
	Brokers  []string                 `yaml:"Brokers"`
	Topics   []*ProducerConfTopicItem `yaml:"Topics"`
}

type ProducerConfTopicItem struct {
	TopicName     string `yaml:"TopicName"`
	TopicAlias    string `yaml:"TopicAlias"`
	KafkaUserName string `yaml:"KafkaUserName"`
	KafkaPassword string `yaml:"KafkaPassword"`
}

type KafKaConf struct {
	ConsumerList []*ConsumerConfItem `yaml:"ConsumerList"`
	ProducerList []*ProducerConfItem `yaml:"ProducerList"`
}

type ESConfigItem struct {
	Name string `yaml:"Name"`
	URL  string `yaml:"URL"`
}

type ESClusterConfig struct {
	ESClusterList []*ESConfigItem `yaml:"ESClusterList"`
	ESIndexList   []*ESIndexItem  `yaml:"ESIndexList"`
}

type ESIndexItem struct {
	Dimension string `yaml:"Dimension"` // ES数据来源组织形式
	IndexName string `yaml:"IndexName"` // ES索引名
}

func DoInitWcConfigWithPath(filePath string) error {
	buf, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, WCConfig)
	if err != nil {
		return err
	}
	marshal, err := json.Marshal(WCConfig)
	fmt.Println(string(marshal))
	return err
}

func init() {
	elem := reflect.TypeOf(WCConfig).Elem()
	for i := 0; i < elem.NumField(); i++ {
		tag := elem.Field(i).Tag.Get("yaml")
		RegisterInitWithConfig(tag, func(i interface{}) {

		}, reflect.ValueOf(elem.Field(i)))
	}
}

package ansilog

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/oblq/sprbox"
)

type KVConfig struct {
	KeyColor       string `yaml:"key_color"`
	KeyMinColWidth int    `yaml:"key_col_width"`
	ValColor       string `yaml:"val_color"`
}

// KVLogger is the ansilog instance type for Key-Value logging.
type KVLogger struct {
	KeyPainter     painter
	KeyMinColWidth int
	ValuePainter   painter
}

func NewKVLogger(configFilePath string, config *KVConfig) *KVLogger {
	if len(configFilePath) > 0 {
		if compsConfigFile, err := ioutil.ReadFile(configFilePath); err != nil {
			panic("wrong config path:" + err.Error())
		} else if err = sprbox.Unmarshal(compsConfigFile, &config); err != nil {
			panic("can't unmarshal config file:" + err.Error())
		}
	}

	kvl := &KVLogger{}
	kvl.KeyPainter = dynamicPainter(color(config.KeyColor))
	kvl.KeyMinColWidth = config.KeyMinColWidth
	kvl.ValuePainter = dynamicPainter(color(config.ValColor))

	return kvl
}

// Go2Box is the https://github.com/oblq/boxes 'boxable' interface implementation.
func (kvl *KVLogger) SpareConfig(configFiles []string) (err error) {
	var config *KVConfig
	if err = sprbox.LoadConfig(&config, configFiles...); err != nil {
		return err
	}

	kvl.KeyPainter = dynamicPainter(color(config.KeyColor))
	kvl.KeyMinColWidth = config.KeyMinColWidth
	kvl.ValuePainter = dynamicPainter(color(config.ValColor))

	return
}

// maxColWidth define the KeyMaxColWidth default value.
var minColWidth = 20

// Print print the key with predefined KeyColor and width
// and the value with the predefined ValueColor.
func (kvl *KVLogger) Print(key interface{}, value interface{}) {
	k, v := kvl.ansify(key, value)
	fmt.Printf("%v%v", k, v)
}

// Println print the key with predefined KeyColor and KeyMaxWidth
// and the value with the predefined ValueColor.
func (kvl *KVLogger) Println(key interface{}, value interface{}) {
	k, v := kvl.ansify(key, value)
	fmt.Printf("%v%v\n", k, v)
}

func (kvl *KVLogger) ansify(key interface{}, value interface{}) (string, string) {
	if kvl.KeyMinColWidth == 0 {
		kvl.KeyMinColWidth = minColWidth
	}

	var k, v string

	if kvl.KeyPainter == nil {
		k = fmt.Sprintf("%-"+strconv.Itoa(kvl.KeyMinColWidth)+"v", key)
	} else {
		k = kvl.KeyPainter(fmt.Sprintf("%-"+strconv.Itoa(kvl.KeyMinColWidth)+"v", key))
	}

	if kvl.ValuePainter != nil {
		v = kvl.ValuePainter(value)
	} else {
		v = fmt.Sprint(value)
	}

	return k, v
}

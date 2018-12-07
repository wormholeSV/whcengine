package config

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var confData = []byte(`
go_version: 1.10.0
version: 1.0.0

testnet: true

private:
  server_secret: 1D1E6531B32D52D
  session_secret: 970D65CC268C9CFFFA4CA929

db:
  user: root
  passwd: 3A836BF12C34D4FEA4C600151A016
  host: 127.0.0.1
  port: 3306
  database: wallet
redis:
  user: root
  passwd: A2BF528202F71857ED5CB86DC1
  host: 127.0.0.1
  port: 6379
  db_num: 1
log:
  filename: app.log
  level: debug
tx:
  mini_output: 0.00000546
`)

func TestInitConfig(t *testing.T) {
	Convey("Given config file", t, func() {
		filename := fmt.Sprintf("conf_test%04d.yml", rand.Intn(9999))
		os.Setenv(ConfEnv, filename)
		path, err := filepath.Abs("./")
		if err != nil {
			panic(err)
		}
		correctPath := filepath.Join(path, filename)
		os.Setenv(ConfTestEnv, correctPath)

		ioutil.WriteFile(filename, confData, 0664)

		Convey("When init configuration", func() {
			config := GetConf()

			Convey("Configuration should resemble default configuration", func() {
				expected := &configuration{}
				expected.GoVersion = "1.10.0"
				expected.Version = "1.0.0"
				expected.TestNet = true
				expected.DB.User = "root"
				expected.DB.Passwd = "3A836BF12C34D4FEA4C600151A016"
				expected.DB.Host = "127.0.0.1"
				expected.DB.Port = 3306
				expected.DB.Database = "wallet"
				expected.Redis.Passwd = "A2BF528202F71857ED5CB86DC1"
				expected.Redis.Host = "127.0.0.1"
				expected.Redis.Port = 6379
				expected.Redis.DbNum = 1
				expected.Log.Filename = "app.log"
				expected.Log.Level = "debug"
				expected.Log.MaxAge = 168
				expected.Tx.MiniOutput = 0.00000546

				So(config, ShouldResemble, expected)
			})
		})

		Reset(func() {
			os.Unsetenv(ConfEnv)
			os.Remove(filename)
		})
	})
}

func Test(t *testing.T) {

}

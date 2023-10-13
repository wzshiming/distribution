package configuration

import (
	"bytes"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/check.v1"
	"gopkg.in/yaml.v2"
)

// Hook up gocheck into the "go test" runner
func Test(t *testing.T) { check.TestingT(t) }

// configStruct is a canonical example configuration, which should map to configYamlV0_1
var configStruct = Configuration{
	Version: "0.1",
	Log: struct {
		AccessLog struct {
			Disabled bool `yaml:"disabled,omitempty"`
		} `yaml:"accesslog,omitempty"`
		Level        Loglevel               `yaml:"level,omitempty"`
		Formatter    string                 `yaml:"formatter,omitempty"`
		Fields       map[string]interface{} `yaml:"fields,omitempty"`
		Hooks        []LogHook              `yaml:"hooks,omitempty"`
		ReportCaller bool                   `yaml:"reportcaller,omitempty"`
	}{
		Level:  "info",
		Fields: map[string]interface{}{"environment": "test"},
	},
	Storage: Storage{
		"somedriver": Parameters{
			"string1": "string-value1",
			"string2": "string-value2",
			"bool1":   true,
			"bool2":   false,
			"nil1":    nil,
			"int1":    42,
			"url1":    "https://foo.example.com",
			"path1":   "/some-path",
		},
	},
	Auth: Auth{
		"silly": Parameters{
			"realm":   "silly",
			"service": "silly",
		},
	},
	Notifications: Notifications{
		Endpoints: []Endpoint{
			{
				Name: "endpoint-1",
				URL:  "http://example.com",
				Headers: http.Header{
					"Authorization": []string{"Bearer <example>"},
				},
				IgnoredMediaTypes: []string{"application/octet-stream"},
				Ignore: Ignore{
					MediaTypes: []string{"application/octet-stream"},
					Actions:    []string{"pull"},
				},
			},
		},
	},
	Catalog: Catalog{
		MaxEntries: 1000,
	},
	HTTP: struct {
		Addr         string        `yaml:"addr,omitempty"`
		Net          string        `yaml:"net,omitempty"`
		Host         string        `yaml:"host,omitempty"`
		Prefix       string        `yaml:"prefix,omitempty"`
		Secret       string        `yaml:"secret,omitempty"`
		RelativeURLs bool          `yaml:"relativeurls,omitempty"`
		DrainTimeout time.Duration `yaml:"draintimeout,omitempty"`
		TLS          struct {
			Certificate  string   `yaml:"certificate,omitempty"`
			Key          string   `yaml:"key,omitempty"`
			ClientCAs    []string `yaml:"clientcas,omitempty"`
			MinimumTLS   string   `yaml:"minimumtls,omitempty"`
			CipherSuites []string `yaml:"ciphersuites,omitempty"`
			LetsEncrypt  struct {
				CacheFile    string   `yaml:"cachefile,omitempty"`
				Email        string   `yaml:"email,omitempty"`
				Hosts        []string `yaml:"hosts,omitempty"`
				DirectoryURL string   `yaml:"directoryurl,omitempty"`
			} `yaml:"letsencrypt,omitempty"`
		} `yaml:"tls,omitempty"`
		Headers http.Header `yaml:"headers,omitempty"`
		Debug   struct {
			Addr       string `yaml:"addr,omitempty"`
			Prometheus struct {
				Enabled bool   `yaml:"enabled,omitempty"`
				Path    string `yaml:"path,omitempty"`
			} `yaml:"prometheus,omitempty"`
		} `yaml:"debug,omitempty"`
		HTTP2 struct {
			Disabled bool `yaml:"disabled,omitempty"`
		} `yaml:"http2,omitempty"`
	}{
		TLS: struct {
			Certificate  string   `yaml:"certificate,omitempty"`
			Key          string   `yaml:"key,omitempty"`
			ClientCAs    []string `yaml:"clientcas,omitempty"`
			MinimumTLS   string   `yaml:"minimumtls,omitempty"`
			CipherSuites []string `yaml:"ciphersuites,omitempty"`
			LetsEncrypt  struct {
				CacheFile    string   `yaml:"cachefile,omitempty"`
				Email        string   `yaml:"email,omitempty"`
				Hosts        []string `yaml:"hosts,omitempty"`
				DirectoryURL string   `yaml:"directoryurl,omitempty"`
			} `yaml:"letsencrypt,omitempty"`
		}{
			ClientCAs: []string{"/path/to/ca.pem"},
		},
		Headers: http.Header{
			"X-Content-Type-Options": []string{"nosniff"},
		},
		HTTP2: struct {
			Disabled bool `yaml:"disabled,omitempty"`
		}{
			Disabled: false,
		},
	},
	Redis: Redis{
		Addr:     "localhost:6379",
		Username: "alice",
		Password: "123456",
		DB:       1,
		Pool: struct {
			MaxIdle     int           `yaml:"maxidle,omitempty"`
			MaxActive   int           `yaml:"maxactive,omitempty"`
			IdleTimeout time.Duration `yaml:"idletimeout,omitempty"`
		}{
			MaxIdle:     16,
			MaxActive:   64,
			IdleTimeout: time.Second * 300,
		},
		DialTimeout:  time.Millisecond * 10,
		ReadTimeout:  time.Millisecond * 10,
		WriteTimeout: time.Millisecond * 10,
	},
}

// configYamlV0_1 is a Version 0.1 yaml document representing configStruct
var configYamlV0_1 = `
version: 0.1
log:
  level: info
  fields:
    environment: test
storage:
  somedriver:
    string1: string-value1
    string2: string-value2
    bool1: true
    bool2: false
    nil1: ~
    int1: 42
    url1: "https://foo.example.com"
    path1: "/some-path"
auth:
  silly:
    realm: silly
    service: silly
notifications:
  endpoints:
    - name: endpoint-1
      url:  http://example.com
      headers:
        Authorization: [Bearer <example>]
      ignoredmediatypes:
        - application/octet-stream
      ignore:
        mediatypes:
           - application/octet-stream
        actions:
           - pull
http:
  clientcas:
    - /path/to/ca.pem
  headers:
    X-Content-Type-Options: [nosniff]
redis:
  addr: localhost:6379
  username: alice
  password: 123456
  db: 1
  pool:
    maxidle: 16
    maxactive: 64
    idletimeout: 300s
  dialtimeout: 10ms
  readtimeout: 10ms
  writetimeout: 10ms
`

// inmemoryConfigYamlV0_1 is a Version 0.1 yaml document specifying an inmemory
// storage driver with no parameters
var inmemoryConfigYamlV0_1 = `
version: 0.1
log:
  level: info
storage: inmemory
auth:
  silly:
    realm: silly
    service: silly
notifications:
  endpoints:
    - name: endpoint-1
      url:  http://example.com
      headers:
        Authorization: [Bearer <example>]
      ignoredmediatypes:
        - application/octet-stream
      ignore:
        mediatypes:
           - application/octet-stream
        actions:
           - pull
http:
  headers:
    X-Content-Type-Options: [nosniff]
`

type ConfigSuite struct {
	expectedConfig *Configuration
}

var _ = check.Suite(new(ConfigSuite))

func (suite *ConfigSuite) SetUpTest(c *check.C) {
	os.Clearenv()
	suite.expectedConfig = copyConfig(configStruct)
}

// TestMarshalRoundtrip validates that configStruct can be marshaled and
// unmarshaled without changing any parameters
func (suite *ConfigSuite) TestMarshalRoundtrip(c *check.C) {
	configBytes, err := yaml.Marshal(suite.expectedConfig)
	c.Assert(err, check.IsNil)
	config, err := Parse(bytes.NewReader(configBytes))
	c.Log(string(configBytes))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseSimple validates that configYamlV0_1 can be parsed into a struct
// matching configStruct
func (suite *ConfigSuite) TestParseSimple(c *check.C) {
	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseInmemory validates that configuration yaml with storage provided as
// a string can be parsed into a Configuration struct with no storage parameters
func (suite *ConfigSuite) TestParseInmemory(c *check.C) {
	suite.expectedConfig.Storage = Storage{"inmemory": Parameters{}}
	suite.expectedConfig.Log.Fields = nil
	suite.expectedConfig.Redis = Redis{}

	config, err := Parse(bytes.NewReader([]byte(inmemoryConfigYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseIncomplete validates that an incomplete yaml configuration cannot
// be parsed without providing environment variables to fill in the missing
// components.
func (suite *ConfigSuite) TestParseIncomplete(c *check.C) {
	incompleteConfigYaml := "version: 0.1"
	_, err := Parse(bytes.NewReader([]byte(incompleteConfigYaml)))
	c.Assert(err, check.NotNil)

	suite.expectedConfig.Log.Fields = nil
	suite.expectedConfig.Storage = Storage{"filesystem": Parameters{"rootdirectory": "/tmp/testroot"}}
	suite.expectedConfig.Auth = Auth{"silly": Parameters{"realm": "silly"}}
	suite.expectedConfig.Notifications = Notifications{}
	suite.expectedConfig.HTTP.Headers = nil
	suite.expectedConfig.Redis = Redis{}

	// Note: this also tests that REGISTRY_STORAGE and
	// REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY can be used together
	os.Setenv("REGISTRY_STORAGE", "filesystem")
	os.Setenv("REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY", "/tmp/testroot")
	os.Setenv("REGISTRY_AUTH", "silly")
	os.Setenv("REGISTRY_AUTH_SILLY_REALM", "silly")

	config, err := Parse(bytes.NewReader([]byte(incompleteConfigYaml)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseWithSameEnvStorage validates that providing environment variables
// that match the given storage type will only include environment-defined
// parameters and remove yaml-defined parameters
func (suite *ConfigSuite) TestParseWithSameEnvStorage(c *check.C) {
	suite.expectedConfig.Storage = Storage{"somedriver": Parameters{"region": "us-east-1"}}

	os.Setenv("REGISTRY_STORAGE", "somedriver")
	os.Setenv("REGISTRY_STORAGE_SOMEDRIVER_REGION", "us-east-1")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseWithDifferentEnvStorageParams validates that providing environment variables that change
// and add to the given storage parameters will change and add parameters to the parsed
// Configuration struct
func (suite *ConfigSuite) TestParseWithDifferentEnvStorageParams(c *check.C) {
	suite.expectedConfig.Storage.setParameter("string1", "us-west-1")
	suite.expectedConfig.Storage.setParameter("bool1", true)
	suite.expectedConfig.Storage.setParameter("newparam", "some Value")

	os.Setenv("REGISTRY_STORAGE_SOMEDRIVER_STRING1", "us-west-1")
	os.Setenv("REGISTRY_STORAGE_SOMEDRIVER_BOOL1", "true")
	os.Setenv("REGISTRY_STORAGE_SOMEDRIVER_NEWPARAM", "some Value")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseWithDifferentEnvStorageType validates that providing an environment variable that
// changes the storage type will be reflected in the parsed Configuration struct
func (suite *ConfigSuite) TestParseWithDifferentEnvStorageType(c *check.C) {
	suite.expectedConfig.Storage = Storage{"inmemory": Parameters{}}

	os.Setenv("REGISTRY_STORAGE", "inmemory")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseWithDifferentEnvStorageTypeAndParams validates that providing an environment variable
// that changes the storage type will be reflected in the parsed Configuration struct and that
// environment storage parameters will also be included
func (suite *ConfigSuite) TestParseWithDifferentEnvStorageTypeAndParams(c *check.C) {
	suite.expectedConfig.Storage = Storage{"filesystem": Parameters{}}
	suite.expectedConfig.Storage.setParameter("rootdirectory", "/tmp/testroot")

	os.Setenv("REGISTRY_STORAGE", "filesystem")
	os.Setenv("REGISTRY_STORAGE_FILESYSTEM_ROOTDIRECTORY", "/tmp/testroot")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseWithSameEnvLoglevel validates that providing an environment variable defining the log
// level to the same as the one provided in the yaml will not change the parsed Configuration struct
func (suite *ConfigSuite) TestParseWithSameEnvLoglevel(c *check.C) {
	os.Setenv("REGISTRY_LOGLEVEL", "info")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseWithDifferentEnvLoglevel validates that providing an environment variable defining the
// log level will override the value provided in the yaml document
func (suite *ConfigSuite) TestParseWithDifferentEnvLoglevel(c *check.C) {
	suite.expectedConfig.Log.Level = "error"

	os.Setenv("REGISTRY_LOG_LEVEL", "error")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseInvalidLoglevel validates that the parser will fail to parse a
// configuration if the loglevel is malformed
func (suite *ConfigSuite) TestParseInvalidLoglevel(c *check.C) {
	invalidConfigYaml := "version: 0.1\nloglevel: derp\nstorage: inmemory"
	_, err := Parse(bytes.NewReader([]byte(invalidConfigYaml)))
	c.Assert(err, check.NotNil)

	os.Setenv("REGISTRY_LOGLEVEL", "derp")

	_, err = Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.NotNil)
}

// TestParseInvalidVersion validates that the parser will fail to parse a newer configuration
// version than the CurrentVersion
func (suite *ConfigSuite) TestParseInvalidVersion(c *check.C) {
	suite.expectedConfig.Version = MajorMinorVersion(CurrentVersion.Major(), CurrentVersion.Minor()+1)
	configBytes, err := yaml.Marshal(suite.expectedConfig)
	c.Assert(err, check.IsNil)
	_, err = Parse(bytes.NewReader(configBytes))
	c.Assert(err, check.NotNil)
}

// TestParseExtraneousVars validates that environment variables referring to
// nonexistent variables don't cause side effects.
func (suite *ConfigSuite) TestParseExtraneousVars(c *check.C) {

	// Environment variables which shouldn't set config items
	os.Setenv("REGISTRY_DUCKS", "quack")
	os.Setenv("REGISTRY_REPORTING_ASDF", "ghjk")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseEnvVarImplicitMaps validates that environment variables can set
// values in maps that don't already exist.
func (suite *ConfigSuite) TestParseEnvVarImplicitMaps(c *check.C) {
	readonly := make(map[string]interface{})
	readonly["enabled"] = true

	maintenance := make(map[string]interface{})
	maintenance["readonly"] = readonly

	suite.expectedConfig.Storage["maintenance"] = maintenance

	os.Setenv("REGISTRY_STORAGE_MAINTENANCE_READONLY_ENABLED", "true")

	config, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
	c.Assert(config, check.DeepEquals, suite.expectedConfig)
}

// TestParseEnvWrongTypeMap validates that incorrectly attempting to unmarshal a
// string over existing map fails.
func (suite *ConfigSuite) TestParseEnvWrongTypeMap(c *check.C) {
	os.Setenv("REGISTRY_STORAGE_SOMEDRIVER", "somestring")

	_, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.NotNil)
}

// TestParseEnvWrongTypeStruct validates that incorrectly attempting to
// unmarshal a string into a struct fails.
func (suite *ConfigSuite) TestParseEnvWrongTypeStruct(c *check.C) {
	os.Setenv("REGISTRY_STORAGE_LOG", "somestring")

	_, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.NotNil)
}

// TestParseEnvWrongTypeSlice validates that incorrectly attempting to
// unmarshal a string into a slice fails.
func (suite *ConfigSuite) TestParseEnvWrongTypeSlice(c *check.C) {
	os.Setenv("REGISTRY_LOG_HOOKS", "somestring")

	_, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.NotNil)
}

// TestParseEnvMany tests several environment variable overrides.
// The result is not checked - the goal of this test is to detect panics
// from misuse of reflection.
func (suite *ConfigSuite) TestParseEnvMany(c *check.C) {
	os.Setenv("REGISTRY_VERSION", "0.1")
	os.Setenv("REGISTRY_LOG_LEVEL", "debug")
	os.Setenv("REGISTRY_LOG_FORMATTER", "json")
	os.Setenv("REGISTRY_LOG_HOOKS", "json")
	os.Setenv("REGISTRY_LOG_FIELDS", "abc: xyz")
	os.Setenv("REGISTRY_LOG_HOOKS", "- type: asdf")
	os.Setenv("REGISTRY_LOGLEVEL", "debug")
	os.Setenv("REGISTRY_STORAGE", "somedriver")
	os.Setenv("REGISTRY_AUTH_PARAMS", "param1: value1")
	os.Setenv("REGISTRY_AUTH_PARAMS_VALUE2", "value2")
	os.Setenv("REGISTRY_AUTH_PARAMS_VALUE2", "value2")

	_, err := Parse(bytes.NewReader([]byte(configYamlV0_1)))
	c.Assert(err, check.IsNil)
}

func checkStructs(c *check.C, t reflect.Type, structsChecked map[string]struct{}) {
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Map || t.Kind() == reflect.Slice {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}
	if _, present := structsChecked[t.String()]; present {
		// Already checked this type
		return
	}

	structsChecked[t.String()] = struct{}{}

	byUpperCase := make(map[string]int)
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		// Check that the yaml tag does not contain an _.
		yamlTag := sf.Tag.Get("yaml")
		if strings.Contains(yamlTag, "_") {
			c.Fatalf("yaml field name includes _ character: %s", yamlTag)
		}
		upper := strings.ToUpper(sf.Name)
		if _, present := byUpperCase[upper]; present {
			c.Fatalf("field name collision in configuration object: %s", sf.Name)
		}
		byUpperCase[upper] = i

		checkStructs(c, sf.Type, structsChecked)
	}
}

// TestValidateConfigStruct makes sure that the config struct has no members
// with yaml tags that would be ambiguous to the environment variable parser.
func (suite *ConfigSuite) TestValidateConfigStruct(c *check.C) {
	structsChecked := make(map[string]struct{})
	checkStructs(c, reflect.TypeOf(Configuration{}), structsChecked)
}

func copyConfig(config Configuration) *Configuration {
	configCopy := new(Configuration)

	configCopy.Version = MajorMinorVersion(config.Version.Major(), config.Version.Minor())
	configCopy.Loglevel = config.Loglevel
	configCopy.Log = config.Log
	configCopy.Catalog = config.Catalog
	configCopy.Log.Fields = make(map[string]interface{}, len(config.Log.Fields))
	for k, v := range config.Log.Fields {
		configCopy.Log.Fields[k] = v
	}

	configCopy.Storage = Storage{config.Storage.Type(): Parameters{}}
	for k, v := range config.Storage.Parameters() {
		configCopy.Storage.setParameter(k, v)
	}

	configCopy.Auth = Auth{config.Auth.Type(): Parameters{}}
	for k, v := range config.Auth.Parameters() {
		configCopy.Auth.setParameter(k, v)
	}

	configCopy.Notifications = Notifications{Endpoints: []Endpoint{}}
	configCopy.Notifications.Endpoints = append(configCopy.Notifications.Endpoints, config.Notifications.Endpoints...)

	configCopy.HTTP.Headers = make(http.Header)
	for k, v := range config.HTTP.Headers {
		configCopy.HTTP.Headers[k] = v
	}

	configCopy.Redis = config.Redis

	return configCopy
}

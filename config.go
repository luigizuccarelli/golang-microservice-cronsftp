package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

/**
 * The json config will always be in the form of parent.key:value pair
 * the reasoning here is that it is easy to maintain and use
 * and also if required can be migtrated to a key value store such as redis
 *
 * Don't dig it - then feel welcome to change it to your hearts content - knock yourself out
 *
 **/

// Config structure - define the json format for our microservice config
type Config struct {
	Level           string     `json:"level"`
	Basedir         string     `json:"base_dir"`
	Port            string     `json:"port"`
	Cache           string     `json:"cache"`
	Cron            string     `json:"cron"`
	SourcePath      string     `json:"source-path"`
	DestinationPath string     `json:"destination-path"`
	Filename        string     `json:"filename"`
	Url             string     `json:"microservice-url"`
	DeleteAll       string     `json:"microservice-deleteall"`
	InsertAll       string     `json:"microservice-insertall"`
	ApiKey          string     `json:"api-key"`
	Sleep           string     `json:"sleep"`
	Sftp            SftpConfig `json:"sftp"`
	MongoDB         Mongodb
}

// Mongodb structure - the base config to connect to mongodb
type Mongodb struct {
	Host           string `json:"host"`
	Port           string `json:"port"`
	DatabaseName   string `json:"name"`
	User           string `json:"user"`
	Password       string `json:"pwd"`
	AdminUser      string `json:"adminuser"`
	AdminPasssword string `json:"adminpwd"`
}

// Sftp structure
type SftpConfig struct {
	Addr   string `json:"addr"`
	User   string `json:"user"`
	Pwd    string `json:"password"`
	Cipher string `json:"cipher"`
}

// declare and set some vars and constants
var start time.Time

const (
	timeformat string = "2006/01/02 03:04:05"
	fmterror   string = "%s \x1b[1;31m[%s] \x1b[0m : %v \n"
	fmtinfo    string = "%s \x1b[1;34m[%s] \x1b[0m  : %s \n"
)

// ReadFile - a utility function that reads the file
// The design here also ensures our test coverage is high
// It takes in a string and returns a byte array and error object
func ReadFile(filename string) ([]byte, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", err)
		return file, err
	} else {
		return file, nil
	}
}

// ParseJson - a utility function that parses a byte array (json)
// The design here also ensures our test coverage is high
// i.e we dont have to run this multiple of times to increase coverage
// It takes in a byte array returns both Config and error objects
func ParseJson(b []byte) (Config, error) {
	var cfg Config
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", err)
		return cfg, err
	} else {
		return cfg, nil
	}
}

// ValidateJson - a utility function that ensures the required fields are valid
// Also helps in ensuring high test coverage
// As the logger can only be configured after we read the config
// I make use of the stdout for error logging
// It takes in a Config object and returns both Config and error objects
func ValidateJson(cfg Config) (Config, error) {
	e := "Config Level is a mandatory field"
	if cfg.Level == "" {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", e)
		return cfg, errors.New(e)
	} else {
		// override if we have envars set
		if os.Getenv("LOG_LEVEL") != "" {
			cfg.Level = os.Getenv("LOG_LEVEL")
		}
	}

	e = "Config Port is a mandatory field"
	if cfg.Port == "" {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", e)
		return cfg, errors.New(e)
	} else {
		// override if we have envars set
		if os.Getenv("SERVER_PORT") != "" {
			cfg.Port = os.Getenv("SERVER_PORT")
		}
	}

	e = "Config MongoDB Host and Port are mandatory fields"
	if cfg.MongoDB.Port == "" || cfg.MongoDB.Host == "" {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", e)
		return cfg, errors.New(e)
	} else {
		// override if we have envars set
		if os.Getenv("MONGODB_HOST") != "" {
			cfg.MongoDB.Host = os.Getenv("MONGODB_HOST")
		}
		if os.Getenv("MONGODB_PORT") != "" {
			cfg.MongoDB.Port = os.Getenv("MONGODB_PORT")
		}
	}

	e = "Config Cron is a mandatory fields"
	if cfg.Cron == "" {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", e)
		return cfg, errors.New(e)
	} else {
		// override if we have envars set
		if os.Getenv("CRON") != "" {
			cfg.Cron = os.Getenv("CRON")
		}
	}

	e = "Sftp Config and child fields are mandatory"
	if cfg.Sftp.User == "" || cfg.Sftp.Pwd == "" || cfg.Sftp.Addr == "" {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", e)
		return cfg, errors.New(e)
	} else {
		// override if we have envars set
		if os.Getenv("SFTP_ADDR") != "" {
			cfg.Sftp.Addr = os.Getenv("SFTP_ADDR")
		}
		if os.Getenv("SFTP_USER") != "" {
			cfg.Sftp.User = os.Getenv("SFTP_USER")
		}
		if os.Getenv("SFTP_PWD") != "" {
			cfg.Sftp.Pwd = os.Getenv("SFTP_PWD")
		}
	}

	e = "File paths and filename are mandatory fields"
	if cfg.SourcePath == "" || cfg.DestinationPath == "" || cfg.Filename == "" {
		fmt.Printf(fmterror, start.Format(timeformat), "ERROR", e)
		return cfg, errors.New(e)
	} else {
		// override if we have envars set
		if os.Getenv("SOURCE_PATH") != "" {
			cfg.SourcePath = os.Getenv("SOURCE_PATH")
		}
		if os.Getenv("DESTINATION_PATH") != "" {
			cfg.DestinationPath = os.Getenv("DESTINATION_PATH")
		}
		if os.Getenv("FILENAME") != "" {
			cfg.Filename = os.Getenv("FILENAME")
		}
	}

	// all checks pass we are good to go
	return cfg, nil
}

// Init - Combine all our utility functions to ensure a valid schema and json data
func Init(filename string) (Config, error) {
	start = time.Now()
	var cfg Config

	file, err := ReadFile(filename)
	if err != nil {
		return cfg, err
	}

	cfg, e := ParseJson(file)
	if e != nil {
		return cfg, e
	}

	cfg, er := ValidateJson(cfg)
	if er != nil {
		return cfg, er
	}

	fmt.Printf(fmtinfo, start.Format(timeformat), "INFO", "Config data read successfully")
	return cfg, nil
}

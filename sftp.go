package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/microlib/simple"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/robfig/cron.v2"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const ()

var (
	logger   simple.Logger
	config   Config
	srcPath  string = "PFT/"
	dstPath  string = "/tmp/"
	filename string = "pubcodes.csv"
	ts       time.Time
	size     int64 = 0
)

func main() {

	// read in the config json
	config, err := Init("config.json")
	//config = cfg
	if err != nil {
		fmt.Printf("%s \x1b[1;31m[%s] \x1b[0m : %v \n", start.Format("2006/01/02 03:04:05"), "ERROR", err)
		os.Exit(0)
	}

	s, _ := strconv.Atoi(config.Sleep)

	cr := cron.New()
	cr.AddFunc(config.Cron,
		func() {
			getFileStatInfo(config)
		})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		cleanup(cr)
		os.Exit(1)
	}()

	cr.Start()

	for {
		logger.Debug(fmt.Sprintf("NOP sleeping for %s seconds", config.Sleep))
		time.Sleep(time.Duration(s) * time.Second)
	}
}

func getFileStatInfo(cfg Config) error {

	logger.Level = cfg.Level
	addr := cfg.Sftp.Addr
	sftpconfig := &ssh.ClientConfig{
		User: cfg.Sftp.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Sftp.Pwd),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		//Ciphers: []string{"3des-cbc", "aes256-cbc", "aes192-cbc", "aes128-cbc"},
		Config: ssh.Config{
			Ciphers: []string{cfg.Sftp.Cipher},
		},
	}
	conn, err := ssh.Dial("tcp", addr, sftpconfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to dial: %s", err.Error()))
		return err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create client: %s", err.Error()))
		return err
	}

	logger.Info(fmt.Sprintf("Connected to sftp server %v", addr))

	// first stat the file to check if it is in the directory
	// and then check if the time stamp or size has changed
	fi, err := client.Stat(cfg.SourcePath + cfg.Filename)
	logger.Info("Getting file meta data")
	if err != nil {
		logger.Warn(fmt.Sprintf("File stat error : %v\n", err.Error()))
		return err
	} else {
		if (fi.Size() != size) || (fi.ModTime() != ts) {
			e := processFileData(client, cfg)
			if e != nil {
				logger.Error(fmt.Sprintf("Error processing : %v\n", err.Error()))
				return err
			}
			ts = fi.ModTime()
			size = fi.Size()
		} else {
			logger.Info("File stats unchanged - no processing required")
		}
	}
	return nil
}

func processFileData(client *sftp.Client, config Config) error {
	srcFile, err := client.Open(config.SourcePath + config.Filename)
	if err != nil {
		logger.Error(fmt.Sprintf("File open error %v", err))
		return err
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(config.DestinationPath + config.Filename)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		return err
	}
	defer dstFile.Close()

	// Copy the file
	srcFile.WriteTo(dstFile)

	// Close connection
	defer client.Close()
	logger.Info("Copied file to local dir")

	// now read and parse the file
	file, err := os.Open(config.DestinationPath + config.Filename)
	if err != nil {
		logger.Error(fmt.Sprintf("%v", err))
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var payload []SchemaInterface

	for scanner.Scan() {
		kv := scanner.Text()
		s := strings.Split(kv, ",")
		val, _ := strconv.ParseInt(s[0], 10, 64)
		data := PubCode{PubId: val, PubData: s[1]}
		si := SchemaInterface{MetaInfo: "none", LastUpdate: time.Now().UnixNano(), Schema: data}
		payload = append(payload, si)
	}

	b, _ := json.MarshalIndent(payload, "", "	")
	logger.Debug(fmt.Sprintf("Payload %s", string(b)))

	// use this for cors
	//res.setHeader("Access-Control-Allow-Origin", "*")
	//res.setHeader("Access-Control-Allow-Methods", "POST")
	//res.setHeader("Access-Control-Allow-Headers", "accept, content-type")

	// set up http object
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := &http.Client{Transport: tr}

	// purge the db collection
	logger.Debug(fmt.Sprintf("Microservice %s", config.Url+config.DeleteAll))
	req, err := http.NewRequest("DELETE", config.Url+config.DeleteAll, nil)
	//req.Header.Set("X-Api-Key", "")
	resp, err := httpClient.Do(req)

	logger.Info(fmt.Sprintf("Connected to host %s", config.Url))

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Delete All %v", err))
		return err
	}
	logger.Debug(fmt.Sprintf("Response from server %s", string(body)))

	// post the data to the microservice endpoint
	req.Header.Set("X-Custom-Header", config.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	// send the payload from the scanner routine
	req, err = http.NewRequest("POST", config.Url+config.InsertAll, bytes.NewBuffer(b))
	req.Header.Set("X-Api-Key", "")
	resp, err = httpClient.Do(req)

	defer resp.Body.Close()
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(fmt.Sprintf("Insert All %v", err))
		return err
	}
	logger.Debug(fmt.Sprintf("Response from server %s", string(body)))

	return nil
}

func cleanup(c *cron.Cron) {
	logger.Warn("Cleanup resources")
	logger.Info("Terminating")
	c.Stop()
}

package main

import "testing"

func TestInit(t *testing.T) {
	// this test should pass
	_, err := Init("tests/config.json")
	if err != nil {
		t.Errorf("Config Init failed %v ", err)
	}

}

func TestFileNotFound(t *testing.T) {
	_, err := ReadFile("tests/config-no-file.json")
	if err == nil {
		t.Errorf("Config Readfile should fail with file not found")
	}
}

func TestParseJson(t *testing.T) {
	var dummy []byte
	_, err := ParseJson(dummy)
	if err == nil {
		t.Errorf("Config ParseJson should fail with unexpected end of JSON string")
	}
}

func TestMissingFields(t *testing.T) {

	var mongo = Mongodb{Host: "testing", Port: "27017"}
	var config = Config{Level: "", Port: "9000", MongoDB: mongo}

	_, err := ValidateJson(config)
	if err == nil {
		t.Errorf("Config ValidateJson should have failed")
	}

	config = Config{Level: "Level", Port: "", MongoDB: mongo}
	_, err = ValidateJson(config)
	if err == nil {
		t.Errorf("Config ValidateJson should have failed")
	}

	mongo = Mongodb{Host: "", Port: "27071"}
	config = Config{Level: "Level", Port: "9000", MongoDB: mongo}
	_, err = ValidateJson(config)
	if err == nil {
		t.Errorf("Config ValidateJson should have failed")
	}

	mongo = Mongodb{Host: "test", Port: ""}
	config = Config{Level: "Level", Port: "9000", MongoDB: mongo}
	_, err = ValidateJson(config)
	if err == nil {
		t.Errorf("Config ValidateJson should have failed")
	}

}

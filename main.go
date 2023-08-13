package main

import (
	"os"
	"path"

	"github.com/glebarez/sqlite"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

type config struct {
	DBRoot       string `yaml:"db_root"`
	MQTTHost     string `yaml:"mqtt_host"`
	MQTTUsername string `yaml:"mqtt_username"`
	MQTTPassword string `yaml:"mqtt_password"`
}

func main() {
	// log.SetLevel(log.DebugLevel)

	cfgFile := "config.yaml"
	if len(os.Args) > 1 {
		cfgFile = os.Args[1]
	}

	c, err := readConfig(cfgFile)
	if err != nil {
		log.Fatalf("failed to read configuration: %s", err.Error())
	}

	dbName := lastFileIn(c.DBRoot)

	log.Infof("Reading file: %s", dbName)

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %s", err.Error())
	}

	m := &MQTT{}
	m.Initialize(c.MQTTHost, c.MQTTUsername, c.MQTTPassword)

	if err := parse(db, m); err != nil {
		log.Fatalf("failed to parse data: %s", err)
	}
}

func readConfig(file string) (*config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var c config

	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func lastFileIn(dir string) string {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("Failed to read dir '%s': %s", dir, err)
	}

	return path.Join(dir, files[len(files)-1].Name())
}

func parse(db *gorm.DB, m *MQTT) error {
	var cars []Car

	if q := db.Preload("Refuelings.FuelType").Find(&cars); q.Error != nil {
		return q.Error
	}

	for i := range cars {
		car := &cars[i]
		log.Printf("car: %s (#%s)", car.Name, car.HEXColor())
		m.InitializeData(car)
	}

	return nil
}

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var measurements = []map[string]string{
	{"ha_value": "name", "name": "Name", "icon": "car-outline", "unit": "", "state_class": ""},
	{"ha_value": "color", "name": "Color", "icon": "palette", "unit": "", "state_class": ""},
	{"ha_value": "mileage", "name": "Mileage", "icon": "counter", "unit": "km", "class": "distance"},
	{"ha_value": "fuel_type", "name": "Fuel type", "icon": "gas-station-outline", "unit": "", "state_class": ""},
	{"ha_value": "volume", "name": "Volume", "icon": "fuel", "unit": "L", "class": "volume"},
	{"ha_value": "price", "name": "Total price", "icon": "currency-eur", "unit": "€", "class": "monetary"},
	{"ha_value": "price_per_unit", "name": "Price per unit", "icon": "currency-eur", "unit": "€/L", "class": "monetary"},
}

type MQTT struct {
	Host     string
	Username string
	Password string

	client mqtt.Client
}

func (m MQTT) Logger() log.FieldLogger {
	return log.WithField("module", "mqtt")
}

// Initialize the Csv m
func (m *MQTT) Initialize(host, username, password string) {
	m.Host = host
	m.Username = username
	m.Password = password

	m.initializeClient()

	m.Logger().Debugln("I am the MQTT module")
	m.Logger().Debugf("  - Host: %s", m.Host)
	m.Logger().Debugf("  - Username: %s", m.Username)
}

func (m *MQTT) initializeClient() {
	m.client = mqtt.NewClient(
		mqtt.NewClientOptions().
			AddBroker(m.Host).
			SetUsername(m.Username).
			SetPassword(m.Password),
	)
}

type deviceStruct struct {
	Model        string   `json:"mdl"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"mf"`
	Identifiers  []string `json:"identifiers"`
}

type payload struct {
	Name              string       `json:"name"`
	ValueTemplate     string       `json:"value_template"`
	UnitOfMeasurement string       `json:"unit_of_measurement,omitempty"`
	Icon              string       `json:"icon"`
	StateTopic        string       `json:"state_topic"`
	ObjectID          string       `json:"object_id"`
	UniqueID          string       `json:"unique_id"`
	Device            deviceStruct `json:"device"`
	StateClass        string       `json:"state_class,omitempty"`
	DeviceClass       string       `json:"device_class,omitempty"`
}

func (m MQTT) broadcastAutoDiscover(car *Car) error {
	identifier := fmt.Sprintf("car_report_%s", car.SaneName())
	identifierLower := strings.ToLower(identifier)

	if m.client == nil {
		return fmt.Errorf("no mqtt client")
	}

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	defer m.client.Disconnect(250)

	for _, measurement := range measurements {
		measurementIdentifier := fmt.Sprintf("%s_%s", identifierLower, measurement["ha_value"])
		device := deviceStruct{
			Model:        car.Name,
			Name:         identifierLower,
			Manufacturer: "CarReport",
			Identifiers:  []string{identifier, identifierLower},
		}

		adTopic := fmt.Sprintf("homeassistant/sensor/%s/%s/config", identifierLower, measurement["ha_value"])

		adPayload := payload{
			Name:              measurement["name"],
			ValueTemplate:     fmt.Sprintf("{{ value_json.%s }}", measurement["ha_value"]),
			UnitOfMeasurement: measurement["unit"],
			Icon:              "mdi:" + measurement["icon"],
			StateTopic:        fmt.Sprintf("homeassistant/sensor/%s/state", identifierLower),
			ObjectID:          measurementIdentifier,
			UniqueID:          measurementIdentifier,
			Device:            device,
		}

		if val, ok := measurement["state_class"]; ok {
			adPayload.StateClass = val
		} else {
			adPayload.StateClass = "measurement"
		}

		if val, ok := measurement["class"]; ok {
			adPayload.DeviceClass = val
		}

		j, err := json.Marshal(adPayload)
		if err != nil {
			return err
		}

		m.Logger().Debugf("Publishing Auto Discovery for %s to %s", measurement["ha_value"], adTopic)
		m.Logger().Debugf("Payload: %s", j)

		if token := m.client.Publish(adTopic, 1, false, j); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}

	return nil
}

func (m MQTT) sendLastMetric(car *Car) error {
	identifier := fmt.Sprintf("car_report_%s", car.SaneName())
	identifierLower := strings.ToLower(identifier)
	adTopic := fmt.Sprintf("homeassistant/sensor/%s/state", identifierLower)

	lastMetric := car.LastRefueling()

	j, err := json.Marshal(lastMetric)
	if err != nil {
		return err
	}

	m.Logger().Infof("Publishing measurement for %s to %s", identifier, adTopic)

	if token := m.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	defer m.client.Disconnect(250)

	if token := m.client.Publish(adTopic, 1, false, j); token.Wait() && token.Error() != nil {
		return token.Error()
	}

	return nil
}

func (m MQTT) InitializeData(car *Car) bool {
	m.Logger().Infof("The MQTT module is initializing the last data for '%s'", car.Name)

	if err := m.broadcastAutoDiscover(car); err != nil {
		m.Logger().Errorf("Error: %s", err)
		return false
	}

	if err := m.sendLastMetric(car); err != nil {
		m.Logger().Errorf("Error: %s", err)
		return false
	}

	return true
}

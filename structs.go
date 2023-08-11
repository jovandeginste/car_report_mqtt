package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Car struct {
	ID             int         `gorm:"column:_id"             json:"-"`
	Name           string      `gorm:"column:car__name"       json:"name"`
	InitialMileage int         `gorm:"column:initial_mileage" json:"initial_mileage"`
	Color          int64       `gorm:"column:color"           json:"-"`
	Refuelings     []Refueling `json:"-"`
}

func (Car) TableName() string {
	return "car"
}

func (c *Car) SaneName() string {
	return strings.ReplaceAll(c.Name, " ", "_")
}

func (c *Car) HEXColor() string {
	col := c.Color

	// Set first 8 bits to 0
	col &= 0x00ffffff

	// Convert to hex
	hexCol := fmt.Sprintf("#%06x", col)

	return hexCol
}

func (c *Car) LastRefueling() CarWithRefueling {
	cwr := CarWithRefueling{
		Car:   *c,
		Color: c.HEXColor(),
	}

	if len(c.Refuelings) == 0 {
		return cwr
	}

	cwr.Refueling = c.Refuelings[len(c.Refuelings)-1]
	cwr.PricePerUnit = cwr.Refueling.PricePerUnit()

	if cwr.Refueling.FuelType != nil {
		cwr.FuelType = cwr.Refueling.FuelType.Name
	}

	return cwr
}

type FuelType struct {
	ID       int    `gorm:"column:_id"`
	Name     string `gorm:"column:fuel_type__name"`
	Category string `gorm:"column:category"`
}

func (FuelType) TableName() string {
	return "fuel_type"
}

type Refueling struct {
	ID         int       `gorm:"column:_id"          json:"-"`
	Date       int64     `gorm:"column:date"         json:"-"`
	Mileage    int       `gorm:"column:mileage"      json:"mileage"`
	Volume     float32   `gorm:"column:volume"       json:"volume"`
	Price      float32   `gorm:"column:price"        json:"price"`
	Partial    bool      `gorm:"column:partial"      json:"partial"`
	Note       string    `gorm:"column:note"         json:"note"`
	FuelTypeID int       `gorm:"column:fuel_type_id" json:"-"`
	CarID      uint      `gorm:"column:car_id"       json:"-"`
	FuelType   *FuelType `json:"-"`
	Car        *Car      `json:"-"`
}

func (r *Refueling) Time() time.Time {
	return time.Unix(r.Date/1000, 0)
}

func (Refueling) TableName() string {
	return "refueling"
}

func (r *Refueling) PricePerUnit() float32 {
	p := float64(r.Price / r.Volume)

	return float32(math.Round(p*1000) / 1000)
}

type CarWithRefueling struct {
	Car
	Refueling
	FuelType     string  `json:"fuel_type"`
	Color        string  `json:"color"`
	PricePerUnit float32 `json:"price_per_unit"`
}

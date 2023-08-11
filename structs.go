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
	InitialMileage int64       `gorm:"column:initial_mileage" json:"initial_mileage"`
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

func (c *Car) LastRefueling() *Refueling {
	if len(c.Refuelings) < 1 {
		return nil
	}

	r := c.Refuelings[len(c.Refuelings)-1]

	return &r
}

func (c *Car) PenultimateRefueling() *Refueling {
	if len(c.Refuelings) < 2 {
		return nil
	}

	r := c.Refuelings[len(c.Refuelings)-2]

	return &r
}

func (c *Car) RefuelingData() CarWithRefueling {
	cwr := CarWithRefueling{
		Car:   *c,
		Color: c.HEXColor(),
	}

	r := c.LastRefueling()
	if r == nil {
		return cwr
	}

	cwr.Refueling = *r
	cwr.PricePerUnit = cwr.Refueling.PricePerUnit()
	cwr.Timestamp = cwr.Refueling.Timestamp()

	if cwr.Refueling.FuelType != nil {
		cwr.FuelType = cwr.Refueling.FuelType.Name
	}

	pr := c.PenultimateRefueling()
	if pr == nil {
		return cwr
	}

	cwr.DeltaMileage = cwr.Mileage - pr.Mileage
	cwr.DeltaTime = (r.Date - pr.Date) / 1000

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
	Mileage    int64     `gorm:"column:mileage"      json:"mileage"`
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

func (r *Refueling) Timestamp() string {
	return r.Time().Format(time.RFC3339)
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
	Timestamp    string  `json:"timestamp"`
	DeltaTime    int64   `json:"delta_time"`
	DeltaMileage int64   `json:"delta_mileage"`
}

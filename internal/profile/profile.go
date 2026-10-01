// Package profile defines the data model for a target's personal details
// and loading it from a YAML config file. Every field is optional.
package profile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Child is a child with a name and birth year.
type Child struct {
	Name string `yaml:"name"`
	Year string `yaml:"year"`
}

// Profile holds all optional personal details about a target.
// Empty fields are simply skipped by the generation engine.
type Profile struct {
	// A. Person
	Name        string   `yaml:"name"`
	Surname     string   `yaml:"surname"`
	FatherName  string   `yaml:"father_name"`
	Nicknames   []string `yaml:"nicknames"`
	BirthDay    string   `yaml:"birth_day"`
	BirthMonth  string   `yaml:"birth_month"`
	BirthYear   string   `yaml:"birth_year"`
	Usernames   []string `yaml:"usernames"`
	LuckyNumber string   `yaml:"lucky_number"`
	Phones      []string `yaml:"phones"`

	// B. Family
	SpouseName    string  `yaml:"spouse_name"`
	SpouseYear    string  `yaml:"spouse_year"`
	Children      []Child `yaml:"children"`
	PetName       string  `yaml:"pet_name"`
	ImportantDate string  `yaml:"important_date"`

	// C. Place
	City      string `yaml:"city"`
	School    string `yaml:"school"`
	Workplace string `yaml:"workplace"`

	// D. Interests
	FootballClub   string   `yaml:"football_club"`
	Motto          string   `yaml:"motto"`
	Brand          string   `yaml:"brand"`
	CarPlate       string   `yaml:"car_plate"`
	Color          string   `yaml:"color"`
	ReligiousWords []string `yaml:"religious_words"`
}

// Load reads a profile from a YAML file.
func Load(path string) (*Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	return &p, nil
}

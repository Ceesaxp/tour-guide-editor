// internal/models/tour.go
package models

import (
	"time"
)

type Tour struct {
	ID          string    `yaml:"id" validate:"required"`
	Name        string    `yaml:"name" validate:"required"`
	Description string    `yaml:"description" validate:"required"`
	StartDate   time.Time `yaml:"start_date" validate:"required"`
	EndDate     time.Time `yaml:"end_date" validate:"required,gtfield=StartDate"`
	Version     string    `yaml:"version" validate:"required"`
	HeroImage   string    `yaml:"hero_image" validate:"required,url"`
	Author      Author    `yaml:"author" validate:"required"`
	Price       int       `yaml:"price" validate:"required,min=0"`
	Nodes       []Node    `yaml:"nodes" validate:"required,dive"`
	Edges       []Edge    `yaml:"edges" validate:"required,dive"`
}

type Author struct {
	Name        string `yaml:"name" validate:"required"`
	ProfileLink string `yaml:"profile_link" validate:"required,url"`
}

type Node struct {
	ID             int         `yaml:"id" validate:"required"`
	Location       Location    `yaml:"location" validate:"required"`
	ShortDesc      string      `yaml:"short_description" validate:"required"`
	Narrative      string      `yaml:"narrative" validate:"required"`
	AudioNarrative string      `yaml:"audio_narrative" validate:"omitempty,url"`
	MediaFiles     []MediaFile `yaml:"media_files" validate:"dive"`
	EntryCondition *Condition  `yaml:"entry_condition" validate:"omitempty"`
	ExitCondition  *Condition  `yaml:"exit_condition" validate:"omitempty"`
}

type Location struct {
	Lat float64 `yaml:"lat" validate:"required,latitude"`
	Lon float64 `yaml:"lon" validate:"required,longitude"`
}

type MediaFile struct {
	Type      string `yaml:"type" validate:"required,oneof=image audio video"`
	URI       string `yaml:"uri" validate:"required,url"`
	SendDelay int    `yaml:"send_delay" validate:"min=0"`
	Narrative string `yaml:"narrative" validate:"omitempty"`
}

type Condition struct {
	Type          string   `yaml:"type" validate:"required,oneof=quiz q&a puzzle"`
	Strict        bool     `yaml:"strict"`
	Question      string   `yaml:"question" validate:"required"`
	CorrectAnswer string   `yaml:"correct_answer" validate:"required"`
	Hints         []string `yaml:"hints" validate:"omitempty"`
	Options       []string `yaml:"options" validate:"required_if=Type quiz,min=2"`
	MediaLink     string   `yaml:"media_link" validate:"omitempty,required_if=Type puzzle,url"`
}

type Edge struct {
	From         int         `yaml:"from" validate:"required"`
	To           int         `yaml:"to" validate:"required,nefield=From"`
	MediaFiles   []MediaFile `yaml:"media_files" validate:"dive"`
	Condition    *Condition  `yaml:"condition" validate:"omitempty"`
	Instructions string      `yaml:"instructions" validate:"omitempty"`
	Silent       bool        `yaml:"silent"`
}

package tms

import "time"

// DatabaseVer tracks the database schema version information.
type DatabaseVer struct {
	// Iteration holds the current iteration of the database schema.
	Iteration int16 `gorm:"unique;not null"`
}

// Strain stores all information about each strain and associated traits.
type Strain struct {
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	StrainID    int    `gorm:"primary_key;auto_increment"`
	ReferenceID int    `gorm:"unique;not null"`
	Name        string `gorm:"not null"`
	Race        string
	Flavors     []Flavor `gorm:"many2many:strain_flavors"` // FIXME: we just started working here.. we need to define what this will actually look like
	Effects     []Effect
}

// Flavor stores strain flavors.
type Flavor struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	FlavorID  int    `gorm:"primary_key;auto_increment"`
	Name      string `gorm:"not null"`
}

// Effect stores strain effects.
type Effect struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	EffectID  int    `gorm:"primary_key;auto_increment"`
	Name      string `gorm:"not null"`
	Category  string `gorm:"not null"`
}

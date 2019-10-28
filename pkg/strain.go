package tms

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"time"
)

var (
	ErrRecordAlreadyExists   = errors.New("the record already exists")
	ErrDatabaseConnectionNil = errors.New("the given database connection is not connected")
	ErrReferenceIDNotSet     = errors.New("the reference ID must be set for this operation")
)

// StrainRepr is the representation of a strain from the JSON input format.
type StrainRepr []struct {
	Name    string   `json:"name"`
	ID      uint     `json:"id"`
	Race    string   `json:"race"`
	Flavors []string `json:"flavors"`
	Effects struct {
		Positive []string `json:"positive"`
		Negative []string `json:"negative"`
		Medical  []string `json:"medical"`
	} `json:"effects"`
}

// ParseStrains populates a StrainRepr from src.
func ParseStrains(src io.Reader) (StrainRepr, error) {
	var s StrainRepr

	b, err := ioutil.ReadAll(src)
	if err != nil {
		return s, errors.Wrap(err, "unable to read strains")
	}

	if err := json.Unmarshal(b, &s); err != nil {
		return s, errors.Wrap(err, "unable to unmarshal strains")
	}
	return s, nil
}

// Strain stores all information about each strain and associated traits, and is used to directly model
// the database schema.
type Strain struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	// StrainID is the unique identifier of the record in the database, not to be confused with the ReferenceID. This
	// field will auto update.
	StrainID uint `gorm:"primary_key;auto_increment"`
	// ReferenceID, distinct from StrainID, is the tracking identifier of the strain.  Specifically this can be
	// used as a reference identifier distinct from the context of the database.
	ReferenceID uint `gorm:"unique;not null"`
	// Name is the name of the strain.
	Name string `gorm:"not null"`
	// Race indicates the strain's genetic makeup.
	Race    string
	Flavors []Flavor `gorm:"many2many:strain_flavors"`
	Effects []Effect `gorm:"many2many:strain_effects"`
	// DB is the database instance
	DB *gorm.DB `gorm:"-"`
}

// FromDBByReferenceID populates the struct with details from the database by searching on the strain ReferenceID.
func (s *Strain) FromDBByReferenceID() error {
	if s.ReferenceID == 0 {
		return ErrReferenceIDNotSet
	}
	tx := s.DB.Begin()
	if err := tx.Error; err != nil {
		return err
		// FIXME: unfinished
	}
	tx.Commit()
	return nil
}

// CreateInDB creates the entry in the database.  An error is returned if the create fails, or if the
// record already exists.
func (s *Strain) CreateInDB() error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	tx := s.DB.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	if !tx.NewRecord(s) {
		return ErrRecordAlreadyExists
	}
	if err := tx.Create(s).Error; err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "unable to create strain record")
	}
	tx.Commit()
	return nil
}

// SaveInDB ensures the record is saved in the database.
func (s *Strain) SaveInDB() error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	tx := s.DB.Begin()
	if err := tx.Model(s).Save(s).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// FlavorsFromDB gets all associated flavors from the database by searching on the strain ReferenceID.
func (s *Strain) FlavorsFromDBByID() ([]Flavor, error) {
	var flavors []Flavor
	if s.ReferenceID == 0 {
		return flavors, ErrReferenceIDNotSet
	}
	if s.DB == nil {
		return flavors, ErrDatabaseConnectionNil
	}
	tx := s.DB.Begin()
	if err := tx.Error; err != nil {
		return flavors, err
	}
	rows, err := tx.Table("strain_flavors").
		Select("flavor.name").
		Joins("JOIN strain ON strain_flavors.strain_strain_id = strain.strain_id").
		Joins("JOIN flavor ON strain_flavors.flavor_flavor_id = flavor.flavor_id").
		Where("strain.reference_id = ?", s.ReferenceID).
		Rows()
	if err != nil {
		tx.Rollback()
		return flavors, err
	}

	for rows.Next() {
		var f Flavor
		if err := rows.Scan(&f.Name); err != nil {
			tx.Rollback()
			return flavors, err
		}
		flavors = append(flavors, f)
	}
	tx.Commit()
	return flavors, nil
}

func (s *Strain) EffectsFromDBByID() ([]Effect, error) {
	var effects []Effect
	if s.ReferenceID == 0 {
		return effects, ErrReferenceIDNotSet
	}
	tx := s.DB.Begin()
	if err := tx.Error; err != nil {
		return effects, err
	}
	rows, err := tx.Table("strain_effects").
		Select("effect.name, effect.category").
		Joins("JOIN strain ON strain_effects.strain_strain_id = strain.strain_id").
		Joins("JOIN effect ON strain_effects.effect_effect_id = effect.effect_id").
		Where("strain.reference_id = ?", s.ReferenceID).
		Rows()
	if err != nil {
		tx.Rollback()
		return effects, err
	}

	for rows.Next() {
		var e Effect
		if err := rows.Scan(&e.Name, &e.Category); err != nil {
			tx.Rollback()
			return effects, err
		}
		effects = append(effects, e)
	}
	tx.Commit()
	return effects, nil
}

// DatabaseVer tracks the database schema version information.
type DatabaseVer struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	VersionID uint `gorm:"primary_key;auto_increment"`
	// Iteration holds the current iteration of the database schema.
	Iteration uint `gorm:"unique;not null"`
}

// Flavor is how the Strain will taste, and is used to directly model the database schema.
type Flavor struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	FlavorID  uint   `gorm:"primary_key;auto_increment"`
	Name      string `gorm:"not null"`
	// DB is the database instance
	DB *gorm.DB `gorm:"-"`
}

// Effect is the observed effects of the Strain on the user, and is used to directly model the database schema.
type Effect struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	EffectID  uint   `gorm:"primary_key;auto_increment"`
	Name      string `gorm:"not null"`
	Category  string `gorm:"not null"`
}

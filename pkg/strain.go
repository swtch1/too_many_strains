package tms

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"time"
)

var (
	ErrRecordAlreadyExists   = errors.New("the record already exists")
	ErrNotExists             = errors.New("the record does not exist")
	ErrDatabaseConnectionNil = errors.New("the given database connection is not connected")
	ErrReferenceIDNotSet     = errors.New("the reference ID must be set for this operation")
)

// Strain stores all information about each strain and associated traits, and is used to directly model
// the database schema.
type Strain struct {
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
	// StrainID is the unique identifier of the record in the database, not to be confused with the ReferenceID. This
	// field will auto update.
	StrainID uint `gorm:"primary_key;auto_increment" json:"-"`
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
	DB *gorm.DB `gorm:"-" json:"-"`
}

// CreateInDB creates the entry in the database.  An error is returned if the create fails, or if the
// record already exists.
func (s *Strain) CreateInDB() error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	if s.ReferenceID == 0 {
		return ErrReferenceIDNotSet
	}

	if !s.DB.NewRecord(s) { // FIXME: testing
		return ErrRecordAlreadyExists
	}

	maxRetries := 3
	if err := s.createWithRetries(maxRetries, 0, nil); err != nil {
		return err
	}
	return nil
}

func (s *Strain) createWithRetries(max, attempts int, err error) error {
	if attempts > 0 {
		log.Tracef("%d failed attempts to create record with ID %d", attempts, s.ReferenceID)
	}
	if attempts > max {
		return err
	}
	err = s.DB.Create(s).Error
	if err != nil {
		attempts++
		err = s.createWithRetries(max, attempts, err)
	}
	return nil
}

// SaveInDB ensures the record is saved in the database.
func (s *Strain) SaveInDB() error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	if err := s.DB.Model(s).Save(s).Error; err != nil {
		return err
	}
	return nil
}

// FromDBByRefID populates the struct with details from the database by searching on the strain ReferenceID.
func (s *Strain) FromDBByRefID() error {
	if s.ReferenceID == 0 {
		return ErrReferenceIDNotSet
	}
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	var c uint
	s.DB.Model(&s).Where(&s).First(&s).Count(&c)
	if c < 1 {
		// don't allow queries for non-existent records
		return ErrNotExists
	}
	flavors, err := s.FlavorsFromDBByRefID()
	if err != nil {
		return err
	}
	effects, err := s.EffectsFromDBByRefID()
	if err != nil {
		return err
	}
	s.Flavors = flavors
	s.Effects = effects

	return nil
}

// FlavorsFromDB gets all associated flavors from the database by searching on the strain ReferenceID.
func (s *Strain) FlavorsFromDBByRefID() ([]Flavor, error) {
	var flavors []Flavor
	if s.ReferenceID == 0 {
		return flavors, ErrReferenceIDNotSet
	}
	if s.DB == nil {
		return flavors, ErrDatabaseConnectionNil
	}
	rows, err := s.DB.Table("strain_flavors").
		Select("flavor.name").
		Joins("JOIN strain ON strain_flavors.strain_strain_id = strain.strain_id").
		Joins("JOIN flavor ON strain_flavors.flavor_flavor_id = flavor.flavor_id").
		Where("strain.reference_id = ?", s.ReferenceID).
		Rows()
	if err != nil {
		return flavors, err
	}

	for rows.Next() {
		var f Flavor
		if err := rows.Scan(&f.Name); err != nil {
			return flavors, err
		}
		flavors = append(flavors, f)
	}
	return flavors, nil
}

func (s *Strain) EffectsFromDBByRefID() ([]Effect, error) {
	var effects []Effect
	if s.ReferenceID == 0 {
		return effects, ErrReferenceIDNotSet
	}
	rows, err := s.DB.Table("strain_effects").
		Select("effect.name, effect.category").
		Joins("JOIN strain ON strain_effects.strain_strain_id = strain.strain_id").
		Joins("JOIN effect ON strain_effects.effect_effect_id = effect.effect_id").
		Where("strain.reference_id = ?", s.ReferenceID).
		Rows()
	if err != nil {
		return effects, err
	}

	for rows.Next() {
		var e Effect
		if err := rows.Scan(&e.Name, &e.Category); err != nil {
			return effects, err
		}
		effects = append(effects, e)
	}
	return effects, nil
}

func (s *Strain) ToStrainRepr() StrainRepr {
	r := StrainRepr{
		Name:    s.Name,
		ID:      s.ReferenceID,
		Race:    s.Race,
		Flavors: []string{},
		Effects: struct {
			Positive []string `json:"positive"`
			Negative []string `json:"negative"`
			Medical  []string `json:"medical"`
		}{},
	}
	for _, f := range s.Flavors {
		r.Flavors = append(r.Flavors, f.Name)
	}
	for _, e := range s.Effects {
		switch e.Category {
		case "positive":
			r.Effects.Positive = append(r.Effects.Positive, e.Name)
		case "negative":
			r.Effects.Negative = append(r.Effects.Negative, e.Name)
		case "medical":
			r.Effects.Medical = append(r.Effects.Medical, e.Name)
		}
	}
	return r
}

type Strains []Strain

// FromDBByFlavorName populates the Strains struct from the database by searching on strain flavor names.
func (s *Strains) FromDBByFlavorName() {
	// TODO: make case insensitive, probably by just converting to lower case for the database and converting back
	// TODO: to Title Case on the return

}

// DatabaseVer tracks the database schema version information.
type DatabaseVer struct {
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
	VersionID uint       `gorm:"primary_key;auto_increment" json:"-"`
	// Iteration holds the current iteration of the database schema.
	Iteration uint `gorm:"unique;not null"`
}

// Flavor is how the Strain will taste, and is used to directly model the database schema.
type Flavor struct {
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
	FlavorID  uint       `gorm:"primary_key;auto_increment" json:"-"`
	Name      string     `gorm:"not null"`
	// DB is the database instance
	DB *gorm.DB `gorm:"-" json:"-"`
}

// Effect is the observed effects of the Strain on the user, and is used to directly model the database schema.
type Effect struct {
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-"`
	EffectID  uint       `gorm:"primary_key;auto_increment" json:"-"`
	Name      string     `gorm:"not null"`
	Category  string     `gorm:"not null"`
}

// StrainRepr is the representation of a strain from the JSON input format.
type StrainRepr struct {
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
func ParseStrains(src io.Reader) ([]StrainRepr, error) {
	var s []StrainRepr

	b, err := ioutil.ReadAll(src)
	if err != nil {
		return s, errors.Wrap(err, "unable to read strains")
	}

	if err := json.Unmarshal(b, &s); err != nil {
		return s, errors.Wrap(err, "unable to unmarshal strains")
	}
	return s, nil
}

func (s *StrainRepr) Write(w io.Writer) {
	b, _ := json.Marshal(s)
	_, _ = w.Write(b)
}

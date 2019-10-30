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
	// CreatedAt is the record creation time, handled by gorm.
	CreatedAt time.Time `json:"-"`
	// CreatedAt is the record update time, handled by gorm.
	UpdatedAt time.Time `json:"-"`
	// DeletedAt is the record deletion time, when using soft deletes in gorm.
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
	Race string
	// Flavors stores all flavors of the strain.
	Flavors []Flavor `gorm:"many2many:strain_flavors"`
	// Effects stores side effects and their category.
	Effects []Effect `gorm:"many2many:strain_effects"`

	// DB is the database instance
	DB *gorm.DB `gorm:"-" json:"-"`
}

// ReplaceInDB creates the entry in the database.  An error is returned if the create fails, or if the
// record already exists.
func (s *Strain) CreateInDB() error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	if s.ReferenceID == 0 {
		return ErrReferenceIDNotSet
	}

	if !s.DB.NewRecord(s) {
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

// FromDBByRefID populates the struct with details from the database by searching on the strain id.
func (s *Strain) FromDBByRefID(id uint) error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	var c uint
	s.ReferenceID = id
	s.DB.Model(&s).Where(&s).First(&s).Count(&c)
	if c < 1 {
		// don't allow queries for non-existent records
		return ErrNotExists
	}
	flavors, err := s.FlavorsFromDBByRefID(id)
	if err != nil {
		return err
	}
	effects, err := s.EffectsFromDBByRefID(id)
	if err != nil {
		return err
	}
	s.Flavors = flavors
	s.Effects = effects

	return nil
}

// FromDBByName populates the struct with details from the database by searching on the strain name.
func (s *Strain) FromDBByName(name string) error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}
	var c uint
	s.Name = name
	s.DB.Model(&s).Where(&s).First(&s).Count(&c)
	if c < 1 {
		return ErrNotExists
	}
	flavors, err := s.FlavorsFromDBByRefID(s.ReferenceID)
	if err != nil {
		return err
	}
	effects, err := s.EffectsFromDBByRefID(s.ReferenceID)
	if err != nil {
		return err
	}
	s.Flavors = flavors
	s.Effects = effects

	return nil
}

// FlavorsFromDB gets all associated flavors from the database by searching on the strain id.
func (s *Strain) FlavorsFromDBByRefID(id uint) (Flavors, error) {
	var flavors []Flavor
	if s.DB == nil {
		return flavors, ErrDatabaseConnectionNil
	}
	rows, err := s.DB.Table("strain_flavors").
		Select("flavor.name").
		Joins("JOIN strain ON strain_flavors.strain_strain_id = strain.strain_id").
		Joins("JOIN flavor ON strain_flavors.flavor_flavor_id = flavor.flavor_id").
		Where("strain.reference_id = ?", id).
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

// EffectsFromDBByRefID gets all associated effects from the database by searching on the strain id.
func (s *Strain) EffectsFromDBByRefID(id uint) (Effects, error) {
	var effects Effects
	rows, err := s.DB.Table("strain_effects").
		Select("effect.name, effect.category").
		Joins("JOIN strain ON strain_effects.strain_strain_id = strain.strain_id").
		Joins("JOIN effect ON strain_effects.effect_effect_id = effect.effect_id").
		Where("strain.reference_id = ?", id).
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

type Strains struct {
	strains []Strain
	DB      *gorm.DB
}

// FromDBByRace populates the struct with all strains from the database by searching on strain race.
func (s *Strains) FromDBByRace(race string) error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}

	rows, err := s.DB.Table("strain").
		Select("strain.name, strain.reference_id").
		Where("strain.race = ?", race).
		Rows()
	if err != nil {
		return errors.Wrapf(err, "unable to get strains by race %s from DB", race)
	}
	for rows.Next() {
		strain := Strain{Race: race, DB: s.DB}
		if err := rows.Scan(&strain.Name, &strain.ReferenceID); err != nil {
			return errors.Wrap(err, "error scanning results for strain search")
		}
		strain.Flavors, err = strain.FlavorsFromDBByRefID(strain.ReferenceID)
		if err != nil {
			return errors.Wrapf(err, "unable to get flavors for strain with reference ID %d", strain.ReferenceID)
		}
		strain.Effects, err = strain.EffectsFromDBByRefID(strain.ReferenceID)
		if err != nil {
			return errors.Wrapf(err, "unable to get effects for strain with reference ID %d", strain.ReferenceID)
		}
		s.strains = append(s.strains, strain)
	}
	return nil
}

// FromDBByFlavor populates the struct with all strains from the database by searching on strain flavor names.
func (s *Strains) FromDBByFlavor(flavor string) error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}

	rows, err := s.DB.Table("strain_flavors").
		Select("strain.name, strain.race, strain.reference_id").
		Joins("JOIN strain ON strain_flavors.strain_strain_id = strain.strain_id").
		Joins("JOIN flavor ON strain_flavors.flavor_flavor_id = flavor.flavor_id").
		Where("flavor.name = ?", flavor).
		Rows()
	if err != nil {
		return errors.Wrapf(err, "unable to get strains by flavor %s from DB", flavor)
	}
	for rows.Next() {
		strain := Strain{DB: s.DB}
		if err := rows.Scan(&strain.Name, &strain.Race, &strain.ReferenceID); err != nil {
			return errors.Wrap(err, "error scanning results for strain search")
		}
		strain.Flavors, err = strain.FlavorsFromDBByRefID(strain.ReferenceID)
		if err != nil {
			return errors.Wrapf(err, "unable to get flavors for strain with reference ID %d", strain.ReferenceID)
		}
		strain.Effects, err = strain.EffectsFromDBByRefID(strain.ReferenceID)
		if err != nil {
			return errors.Wrapf(err, "unable to get effects for strain with reference ID %d", strain.ReferenceID)
		}
		s.strains = append(s.strains, strain)
	}
	return nil
}

// FromDBByEffect populates the struct with all strains from the database by searching on strain effect name and category.
func (s *Strains) FromDBByEffect(effect string) error {
	if s.DB == nil {
		return ErrDatabaseConnectionNil
	}

	rows, err := s.DB.Table("strain_effects").
		Select("strain.name, strain.race, strain.reference_id").
		Joins("JOIN strain ON strain_effects.strain_strain_id = strain.strain_id").
		Joins("JOIN effect ON strain_effects.effect_effect_id = effect.effect_id").
		Where("effect.name = ?", effect).
		Rows()
	if err != nil {
		return errors.Wrapf(err, "unable to get strains by effect %s from DB", effect)
	}
	for rows.Next() {
		strain := Strain{DB: s.DB}
		if err := rows.Scan(&strain.Name, &strain.Race, &strain.ReferenceID); err != nil {
			return errors.Wrap(err, "error scanning results for strain search")
		}
		strain.Flavors, err = strain.FlavorsFromDBByRefID(strain.ReferenceID)
		if err != nil {
			return errors.Wrapf(err, "unable to get flavors for strain with reference ID %d", strain.ReferenceID)
		}
		strain.Effects, err = strain.EffectsFromDBByRefID(strain.ReferenceID)
		if err != nil {
			return errors.Wrapf(err, "unable to get effects for strain with reference ID %d", strain.ReferenceID)
		}
		s.strains = append(s.strains, strain)
	}
	return nil
}

func (s *Strains) ToStrainRepr() StrainReprs {
	var reprs []StrainRepr
	for _, strain := range s.strains {
		reprs = append(reprs, strain.ToStrainRepr())
	}
	return reprs
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

type Flavors []Flavor

// Difference returns source Flavors minus flavors.  This comparison is done based on the flavor name only.
func (f *Flavors) Difference(flavors Flavors) Flavors {
	var result Flavors
	for _, src := range *f {
		var match bool
		for _, cmp := range flavors {
			if cmp.Name == src.Name {
				match = true
			}
		}
		if !match {
			result = append(result, src)
		}
	}
	return result
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

type Effects []Effect

// Difference returns source Effects minus effects.  This comparison is done based on the effect name and category only.
func (e *Effects) Difference(effects Effects) Effects {
	var result Effects
	for _, src := range *e {
		var match bool
		for _, cmp := range effects {
			if cmp.Name == src.Name && cmp.Category == src.Category {
				match = true
			}
		}
		if !match {
			result = append(result, src)
		}
	}
	return result
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

	DB *gorm.DB `json:"-"`
}

// CreateInDB will create the strain record in the database.  An error is returned if the strain ID already exists.
func (rs *StrainRepr) CreateInDB() error {
	if rs.DB == nil {
		return ErrDatabaseConnectionNil
	}
	var count int
	rs.DB.Select("strain").Where("strain.reference_id = ?", rs.ID).Count(&count)
	if count > 0 {
		return ErrRecordAlreadyExists
	}
	return rs.ReplaceInDB()
}

// ReplaceInDB will create or replace the strain record in the database.
func (rs *StrainRepr) ReplaceInDB() error {
	if rs.DB == nil {
		return ErrDatabaseConnectionNil
	}

	var flavors []Flavor
	for _, flavor := range rs.Flavors {
		f := Flavor{Name: flavor}
		rs.DB.Model(&f).Where(&f).FirstOrCreate(&f)
		flavors = append(flavors, f)
	}

	var effects []Effect
	for _, effect := range rs.Effects.Positive {
		e := Effect{Name: effect, Category: "positive"}
		rs.DB.Model(&e).Where(&e).FirstOrCreate(&e)
		effects = append(effects, e)
	}
	for _, effect := range rs.Effects.Negative {
		e := Effect{Name: effect, Category: "negative"}
		rs.DB.Model(&e).Where(&e).FirstOrCreate(&e)
		effects = append(effects, e)
	}
	for _, effect := range rs.Effects.Medical {
		e := Effect{Name: effect, Category: "medical"}
		rs.DB.Model(&e).Where(&e).FirstOrCreate(&e)
		effects = append(effects, e)
	}

	var s Strain
	s.DB = rs.DB
	s.ReferenceID = rs.ID
	rs.DB.Model(&s).Where(&s).FirstOrCreate(&s)

	// remove flavors that are present in the DB but not in our object
	flavorsFromDB, err := s.FlavorsFromDBByRefID(rs.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get flavors from database")
	}
	for _, superfluousFlavor := range flavorsFromDB.Difference(flavors) {
		_, err := rs.DB.Table("strain_flavors").
			Delete("strain_flavors").
			Joins("JOIN strain ON strain_flavors.strain_strain_id = strain.strain_id").
			Joins("JOIN flavor ON strain_flavors.flavor_flavor_id = flavor.flavor_id").
			Where("flavor.name = ? AND strain.reference_id = ?", superfluousFlavor.Name, s.ReferenceID).
			Rows()
		if err != nil {
			return errors.Wrapf(err, "unable to delete superfluous strain_flavors for ID %d", rs.ID)
		}
	}

	// remove effects that are present in the DB but not in our object
	effectsFromDB, err := s.EffectsFromDBByRefID(rs.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get effects from database")
	}
	for _, superfluousEffect := range effectsFromDB.Difference(effects) {
		_, err := rs.DB.Table("strain_effects").
			Delete("strain_effects").
			Joins("JOIN strain ON strain_effects.strain_strain_id = strain.strain_id").
			Joins("JOIN effect ON strain_effects.effect_effect_id = effect.effect_id").
			Where("effect.name = ? AND strain.reference_id = ?", superfluousEffect.Name, s.ReferenceID).
			Rows()
		if err != nil {
			return errors.Wrapf(err, "unable to delete superfluous strain_effects for ID %d", rs.ID)
		}
	}

	s.Name = rs.Name
	s.Race = rs.Race
	s.Flavors = flavors
	s.Effects = effects

	log.Debugf("updating record for strain %s with ID %d", s.Name, rs.ID)
	if err := rs.DB.Model(&s).Save(&s).Error; err != nil {
		return errors.Wrapf(err, "unable to save record for strain with ID %d", rs.ID)
	}
	return nil
}

// ParseStrain populates a StrainRepr from src.
func ParseStrain(src io.Reader) (StrainRepr, error) {
	var r StrainRepr
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return r, errors.Wrap(err, "unable to read strains")
	}

	if err := json.Unmarshal(b, &r); err != nil {
		return r, errors.Wrap(err, "unable to unmarshal strains")
	}
	return r, nil
}

// ParseStrains populates a StrainReprs from src.
func ParseStrains(src io.Reader) (StrainReprs, error) {
	var r StrainReprs
	b, err := ioutil.ReadAll(src)
	if err != nil {
		return r, errors.Wrap(err, "unable to read strains")
	}

	if err := json.Unmarshal(b, &r); err != nil {
		return r, errors.Wrap(err, "unable to unmarshal strains")
	}
	return r, nil
}

func (rs *StrainRepr) Write(w io.Writer) {
	b, _ := json.Marshal(rs)
	_, _ = w.Write(b)
}

type StrainReprs []StrainRepr

func (rs *StrainReprs) ToJson() ([]byte, error) {
	b, err := json.Marshal(rs)
	if err != nil {
		return []byte{}, errors.Wrap(err, "failed to unmarshal StrainReprs")
	}
	return b, err
}

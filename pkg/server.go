package tms

import (
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

var ErrStrainIdMustBeInteger = errors.New("strain ID must be an integer")

type Server struct {
	// Port is the port where the server will listen.
	Port int32
	// DB is the database instance
	DB *gorm.DB
}

func (s *Server) ListenAndServe() error {
	r := mux.NewRouter()
	r.HandleFunc("/api/strains/", s.CreateStrainHandler).Methods("POST")
	r.HandleFunc("/api/strains/id/{id}", s.StrainByIDHandler).Methods("GET", "PUT", "DELETE")
	r.HandleFunc("/api/strains/name/{name}", s.StrainByNameHandler).Methods("GET")
	r.HandleFunc("/api/strains/race/{race}", s.StrainByRaceHandler).Methods("GET")
	r.HandleFunc("/api/strains/effect/{effect}", s.StrainByEffectHandler).Methods("GET")
	r.HandleFunc("/api/strains/flavor/{flavor}", s.StrainByFlavorHandler).Methods("GET")
	r.Use(LogInboundRequestMw)

	http.Handle("/", r)
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      r,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	return httpSrv.ListenAndServe()
}

func (s *Server) StrainByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strain := s.newStrain()

	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Debugf("request for strain with non-integer ID %s", vars["id"])
		_, _ = fmt.Fprintf(w, "%s\n", ErrStrainIdMustBeInteger)
		return
	}

	switch r.Method {
	case http.MethodGet:
		err := strain.FromDBByRefID(uint(id))
		if err == ErrNotExists {
			w.WriteHeader(http.StatusNotFound)
			log.WithError(err).Debugf("request for strain with ID %d, strain not found", id)
			_, _ = fmt.Fprintf(w, "404 strain not found\n")
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("could not get strain with ID %d", id)
			_, _ = fmt.Fprintf(w, "%s\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		repr := strain.ToStrainRepr()
		repr.Write(w)
		_, _ = fmt.Fprintf(w, "\n")

	case http.MethodPut:

	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 page not found\n")
	}
}

func (s *Server) CreateStrainHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		repr, err := ParseStrain(r.Body)
		if err != nil {
			panic(err)
		}

		repr.DB = s.DB
		err = repr.ReplaceInDB()
		if err == ErrRecordAlreadyExists {
			w.WriteHeader(http.StatusConflict)
			_, _ = fmt.Fprintf(w, "409 strain with ID %d already exists\n", repr.ID)
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprintf(w, "%s\n", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		// TODO: write https instead if they are using TLS
		_, _ = fmt.Fprintf(w, `{"link"":"http://%s/api/strains/id/%d"}`, r.Host, repr.ID)

	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 page not found\n")
	}
}

func (s *Server) StrainByNameHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strain := s.newStrain()

	switch r.Method {
	case http.MethodGet:
		err := strain.FromDBByName(vars["name"])
		if err == ErrNotExists {
			w.WriteHeader(http.StatusNotFound)
			log.WithError(err).Debugf("request for strain with name %s, strain not found", vars["name"])
			_, _ = fmt.Fprintf(w, "404 strain not found\n")
			return
		} else if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("could not get strain with name %s", vars["name"])
			_, _ = fmt.Fprintf(w, "%s\n", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		repr := strain.ToStrainRepr()
		repr.Write(w)
		_, _ = fmt.Fprintf(w, "\n")

	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 page not found\n")
	}
}

func (s *Server) StrainByRaceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strains := s.newStrains()

	switch r.Method {
	case http.MethodGet:
		if err := strains.FromDBByRace(vars["race"]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("could not get strains by race for race %s", vars["race"])
		}
		w.WriteHeader(http.StatusOK)
		strainReprs := strains.ToStrainRepr()
		b, err := strainReprs.ToJson()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("failed to unmarshal getting strain by race")
		}
		_, _ = fmt.Fprintf(w, string(b))
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 page not found\n")
	}
}

func (s *Server) StrainByFlavorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strains := s.newStrains()

	switch r.Method {
	case http.MethodGet:
		if err := strains.FromDBByFlavor(vars["flavor"]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("could not get strains by flavor for flavor %s", vars["flavor"])
		}
		w.WriteHeader(http.StatusOK)
		strainReprs := strains.ToStrainRepr()
		b, err := strainReprs.ToJson()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("failed to unmarshal getting strain by flavor")
		}
		_, _ = fmt.Fprintf(w, string(b))
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 page not found\n")
	}
}

func (s *Server) StrainByEffectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	strains := s.newStrains()

	switch r.Method {
	case http.MethodGet:
		if err := strains.FromDBByEffect(vars["effect"]); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("could not get strains by effect for effect %s", vars["effect"])
		}
		w.WriteHeader(http.StatusOK)
		strainReprs := strains.ToStrainRepr()
		b, err := strainReprs.ToJson()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.WithError(err).Errorf("failed to unmarshal getting strain by effect")
		}
		_, _ = fmt.Fprintf(w, string(b))
	default:
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, "404 page not found\n")
	}
}

func (s *Server) newStrain() Strain {
	return Strain{DB: s.DB}
}

func (s *Server) newStrains() Strains {
	return Strains{DB: s.DB}
}

func LogInboundRequestMw(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Tracef("%s request from addr %s", r.Method, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

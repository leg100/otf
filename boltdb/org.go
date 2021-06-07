package boltdb

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	bolt "go.etcd.io/bbolt"
)

var _ ots.OrganizationService = (*OrganizationService)(nil)

type OrganizationService struct {
	*bolt.DB
}

func NewOrganizationService(db *bolt.DB) *OrganizationService {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Organizations"))
		if err != nil {
			panic(err.Error())
		}
		return nil
	})
	return &OrganizationService{DB: db}
}

func (s OrganizationService) CreateOrganization(name string, opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
	org := &ots.Organization{}

	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Organizations"))

		if b.Get([]byte(name)) != nil {
			return fmt.Errorf("already exists")
		}

		org, err := ots.NewOrganizationFromOptions(opts)
		if err != nil {
			return err
		}

		buf, err := json.Marshal(org)
		if err != nil {
			return err
		}

		if err := b.Put([]byte(name), []byte(buf)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return org, err
}

func (s OrganizationService) GetOrganization(name string) (*ots.Organization, error) {
	org := &ots.Organization{}

	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Organizations"))

		v := b.Get([]byte(name))

		if err := json.Unmarshal(v, org); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return org, err
}

func (s OrganizationService) ListOrganizations() ([]*ots.Organization, error) {
	var orgs []*ots.Organization

	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Organizations"))

		b.ForEach(func(k, v []byte) error {
			o := &ots.Organization{}
			if err := json.Unmarshal(v, o); err != nil {
				return err
			}
			orgs = append(orgs, o)
			return nil
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	return orgs, err
}

func (s OrganizationService) UpdateOrganization(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
	org := &ots.Organization{}

	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Organizations"))

		v := b.Get([]byte(name))
		if v == nil {
			return errors.New("not found")
		}

		if err := json.Unmarshal(v, org); err != nil {
			return fmt.Errorf("unable to decode: %w", err)
		}

		if err := ots.UpdateOrganizationFromOptions(org, opts); err != nil {
			return err
		}

		buf, err := json.Marshal(org)
		if err != nil {
			return err
		}

		if err := b.Put([]byte(name), []byte(buf)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return org, err
}

func (s OrganizationService) DeleteOrganization(name string) error {
	err := s.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Organizations"))

		if v := b.Get([]byte(name)); v == nil {
			return errors.New("not found")
		}

		if err := b.Delete([]byte(name)); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return err
}

// GetEntitlements currently shows all entitlements as disabled for an org.
func (s OrganizationService) GetEntitlements(name string) (*ots.Entitlements, error) {
	err := s.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Organizations"))

		if v := b.Get([]byte(name)); v == nil {
			return errors.New("not found")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ots.NewEntitlements(name), nil
}

package types

import (
	crud "github.com/iov-one/cosmos-sdk-crud/types"
)

type Store interface {
	Create(o crud.Object) error
	Read(primaryKey []byte, o crud.Object) error
	Update(o crud.Object) error
	Delete(primaryKey []byte) error
}

type CantFailStore struct {
	store Store
}

func (p CantFailStore) Create(o crud.Object) {
	if err := p.store.Create(o); err != nil {
		panic(err)
	}
}

func (p CantFailStore) Read(primaryKey []byte, o crud.Object) {
	if err := p.store.Read(primaryKey, o); err != nil {
		panic(err)
	}
}

func (p CantFailStore) Update(o crud.Object) {
	if err := p.store.Update(o); err != nil {
		panic(err)
	}
}

func (p CantFailStore) Delete(primaryKey []byte) {
	if err := p.store.Delete(primaryKey); err != nil {
		panic(err)
	}
}

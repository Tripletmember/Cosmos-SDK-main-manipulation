package types

type Store interface {
	Create(o Object) error
	Read(primaryKey []byte, o Object) error
	Update(o Object) error
	Delete(primaryKey []byte) error
}

type CantFailStore struct {
	store Store
}

func (p CantFailStore) Create(o Object) {
	if err := p.store.Create(o); err != nil {
		panic(err)
	}
}

func (p CantFailStore) Read(primaryKey []byte, o Object) {
	if err := p.store.Read(primaryKey, o); err != nil {
		panic(err)
	}
}

func (p CantFailStore) Update(o Object) {
	if err := p.store.Update(o); err != nil {
		panic(err)
	}
}

func (p CantFailStore) Delete(primaryKey []byte) {
	if err := p.store.Delete(primaryKey); err != nil {
		panic(err)
	}
}

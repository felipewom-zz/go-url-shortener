package factory

import (
	"github.com/felipewom/go-url-shortener/internal/store"
	"math/rand"
	"net/url"
	"time"
)

// Generator the type to generate keys(short urls)
type Generator func() string

// DefaultGenerator is the defautl url generator
var DefaultGenerator = func() string {
	return randSeq(10)
}

// Factory is responsible to generate keys(short urls)
type Factory struct {
	dataStore store.Store
	generator Generator
}

// NewFactory receives a generator and a store and returns a new url Factory.
func NewFactory(generator Generator, st store.Store) *Factory {
	return &Factory{
		dataStore: st,
		generator: generator,
	}
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Gen generates the key.
func (f *Factory) Gen(uri string) (key string, err error) {
	// we don't return the parsed url because #hash are converted to uri-compatible
	// and we don't want to encode/decode all the time, there is no need for that,
	// we save the url as the user expects if the uri validation passed.
	_, err = url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}

	key = f.generator()
	// Make sure that the key is unique
	for {
		if v := f.dataStore.Get(key); v == "" {
			break
		}
		key = f.generator()
	}

	return key, nil
}

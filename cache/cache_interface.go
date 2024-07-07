package cache

type ICache interface {
	Set(key string, value string) error
	Get(key string) (string, bool)
	Has(key string) bool
	Delete(key string) error
}

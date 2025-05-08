package sqlite

import "gorm.io/gorm"

// WithTx is a decorator that runs a function in a transaction.
// It returns the result of the function and an error if the transaction fails.
// It also handles the rollback of the transaction if the function returns an error.
func WithTx[T any](db *gorm.DB, fn func(tx *gorm.DB) (T, error)) (T, error) {
	var result T
	err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		result, err = fn(tx)
		return err
	})
	return result, err
}

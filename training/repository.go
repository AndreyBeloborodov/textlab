
package training

import "errors"

var ErrNotFound = errors.New("training not found")

type Repository interface {
    Create(training Training) (Training, error)
    Get(id int) (Training, error)
    Update(training Training) error
    GetByModelID(modelID int) ([]Training, error)
}

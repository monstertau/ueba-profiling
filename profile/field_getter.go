package profile

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"ueba-profiling/view"
)

const (
	fieldTypeOriginal = "original"
)

type (
	FieldGetter interface {
		// Get returns the Value extracted from the json msg
		Get(msg []byte) (interface{}, bool)

		// GetTargetFieldName returns the field name
		GetTargetFieldName() string
	}
	originalFieldGetter struct {
		fieldName string
	}
)

func fieldGetterFactory(obj *view.Object) (FieldGetter, error) {
	if len(obj.Name) == 0 {
		return nil, errors.New("required field_name")
	}
	switch obj.Type {
	case fieldTypeOriginal:
		return &originalFieldGetter{
			fieldName: obj.Name,
		}, nil

	}
	return nil, errors.Errorf("unsupported type '%s'", obj.Type)
}

func (g *originalFieldGetter) GetTargetFieldName() string {
	return g.fieldName
}

func (g *originalFieldGetter) Get(msg []byte) (interface{}, bool) {
	if !gjson.ValidBytes(msg) {
		return nil, false
	}
	val := gjson.GetBytes(msg, g.fieldName).Value()
	if val == nil {
		return val, false
	}
	return val, true
}

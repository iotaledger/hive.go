package serix

import (
	"reflect"
	"sync"

	"github.com/iotaledger/hive.go/ierrors"
)

type validators struct {
	syntacticValidator reflect.Value
}

func parseValidatorFunc(validatorFn interface{}) (reflect.Value, error) {
	if validatorFn == nil {
		return reflect.Value{}, nil
	}
	funcValue := reflect.ValueOf(validatorFn)
	if !funcValue.IsValid() || funcValue.IsZero() {
		return reflect.Value{}, nil
	}
	if funcValue.Kind() != reflect.Func {
		return reflect.Value{}, ierrors.Errorf(
			"validator must be a function, got %T(%s)", validatorFn, funcValue.Kind(),
		)
	}
	funcType := funcValue.Type()
	if funcType.NumIn() != 2 {
		return reflect.Value{}, ierrors.New("validator func must have two arguments")
	}
	firstArgType := funcType.In(0)
	if firstArgType != ctxType {
		return reflect.Value{}, ierrors.New("validator func's first argument must be context")
	}
	if funcType.NumOut() != 1 {
		return reflect.Value{}, ierrors.Errorf("validator func must have one return value, got %d", funcType.NumOut())
	}
	returnType := funcType.Out(0)
	if returnType != errorType {
		return reflect.Value{}, ierrors.Errorf("validator func must have 'error' return type, got %s", returnType)
	}

	return funcValue, nil
}

func checkSyntacticValidatorSignature(objectType reflect.Type, funcValue reflect.Value) error {
	funcType := funcValue.Type()
	argumentType := funcType.In(1)
	if argumentType != objectType {
		return ierrors.Errorf(
			"syntacticValidatorFn's argument must have the same type as the object it was registered for, "+
				"objectType=%s, argumentType=%s",
			objectType, argumentType,
		)
	}

	return nil
}

type validatorsRegistry struct {
	// the registered validators for the known objects
	registryMutex sync.RWMutex
	registry      map[reflect.Type]validators
}

func newValidatorsRegistry() *validatorsRegistry {
	return &validatorsRegistry{
		registry: make(map[reflect.Type]validators),
	}
}

func (r *validatorsRegistry) Has(objType reflect.Type) bool {
	_, exists := r.Get(objType)

	return exists
}

func (r *validatorsRegistry) Get(objType reflect.Type) (validators, bool) {
	r.registryMutex.RLock()
	defer r.registryMutex.RUnlock()

	vldtrs, exists := r.registry[objType]

	return vldtrs, exists
}

func (r *validatorsRegistry) RegisterValidator(obj any, syntacticValidatorFn interface{}) error {
	objType := reflect.TypeOf(obj)
	if objType == nil {
		return ierrors.New("'obj' is a nil interface, it needs to be a valid type")
	}

	r.registryMutex.Lock()
	defer r.registryMutex.Unlock()

	if _, exists := r.registry[objType]; exists {
		return ierrors.Errorf("validator for object type %s is already registered", objType)
	}

	syntacticValidatorValue, err := parseValidatorFunc(syntacticValidatorFn)
	if err != nil {
		return ierrors.Wrap(err, "failed to parse syntacticValidatorFn")
	}

	vldtrs := validators{}

	if syntacticValidatorValue.IsValid() {
		if err := checkSyntacticValidatorSignature(objType, syntacticValidatorValue); err != nil {
			return ierrors.WithStack(err)
		}
		vldtrs.syntacticValidator = syntacticValidatorValue
	}

	r.registry[objType] = vldtrs

	return nil
}

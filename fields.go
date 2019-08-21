package fields

import (
	"fmt"
	"io"
	"net/http"
)

// BodyDecoder decodes the body of a request. Implementations of this interface should return an error if the body differs from what's expected.
type BodyDecoder interface {
	DecodeBody(body io.Reader) error
}

// PathVarsDecoder decodes path variables from the request into a map for further validation. Implementations should not perform any validations of the parameters. This should be done in a PathVarValidator implementation.
type PathVarsDecoder interface {
	DecodePathVars(r *http.Request) map[string]string
}

// PathVarsDecoderFunc is an adapter to allow the use of ordinary functions as path variable decoders. If f is a function with the appropriate signature, PathVarsDecoderFunc(f) is a PathVarsDecoder that calls f.
type PathVarsDecoderFunc func(r *http.Request) map[string]string

// DecodePathVars calls f(r)
func (f PathVarsDecoderFunc) DecodePathVars(r *http.Request) map[string]string {
	return f(r)
}

// PathVarValidator validates path variables. Implementations should return an error if a variable has an unexpected format.
type PathVarValidator interface {
	ValidatePathVar(key, value string) error
}

// PathVarValidatorFunc is an adapter to allow the use of ordinary functions as path variable validator. If f is a function with the appropriate signature, PathVarValidatorFunc(f) is a PathVarValidator that calls f.
type PathVarValidatorFunc func(key, value string) error

// ValidatePathVar calls f(key, value)
func (f PathVarValidatorFunc) ValidatePathVar(key, value string) error {
	return f(key, value)
}

// ExpectChain is used for validating a http.Request. Most functions can be chained together which allows a compact description of expected fields. Cains should en with a call to Parse.
type ExpectChain struct {
	pd    PathVarsDecoder
	pv    PathVarValidator
	bd    BodyDecoder
	expPV []string
}

// Expect is a convenience function for starting a chain
func Expect() *ExpectChain {
	return &ExpectChain{}
}

// WithPathVars defines how path vars should be handled.
func (e *ExpectChain) WithPathVars(d PathVarsDecoder, v PathVarValidator) *ExpectChain {
	e.pd = d
	e.pv = v
	return e
}

// Body sets an expectation for a body. Implementations of BodyDecoder is responsible for retaining data parsed from the body.
func (e *ExpectChain) Body(d BodyDecoder) *ExpectChain {
	e.bd = d
	return e
}

// PathVar sets an expectation for a path variable. Must be used together with WithPathVars.
func (e *ExpectChain) PathVar(key string) *ExpectChain {
	e.expPV = append(e.expPV, key)
	return e
}

// Parse ends a chain and verifies that all expected fields are set.
func (e *ExpectChain) Parse(r *http.Request) error {
	// Path vars
	if len(e.expPV) > 0 && e.pd == nil || e.pv == nil {
		panic("you need to set a PathVarsDecoder and a PathVarValidator (see WithPathVars) to use PathVar")
	}

	vars := e.pd.DecodePathVars(r)
	for _, key := range e.expPV {
		v, ok := vars[key]
		if !ok {
			return fmt.Errorf("expected path var: %s", key)
		}

		if err := e.pv.ValidatePathVar(key, v); err != nil {
			return err
		}
	}

	// Body
	if e.bd != nil {
		if err := e.bd.DecodeBody(r.Body); err != nil {
			return err
		}
	}

	return nil
}

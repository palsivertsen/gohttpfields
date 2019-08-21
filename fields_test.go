package fields

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

//go:generate mockgen -source fields.go -destination fields_mock.go -package fields

func TestExpectChain_Parse(t *testing.T) {
	tests := map[string]struct {
		body     interface{}
		pathVars map[string]string
		r        *http.Request
	}{
		"string array body": {
			body: []string{"hello", "world!"},
			r:    httptest.NewRequest(http.MethodGet, "http://", strings.NewReader(`["hello","world!"]`)),
		},
		"path var: foo:bar": {
			pathVars: map[string]string{
				"foo": "bar",
			},
			r: httptest.NewRequest(http.MethodGet, "http://", nil),
		},
	}
	for name, tt := range tests {
		tt := tt
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPathVarsDecoder := NewMockPathVarsDecoder(ctrl)
			mockPathVarsDecoder.EXPECT().DecodePathVars(tt.r).Return(tt.pathVars)

			mockPathVarValidator := NewMockPathVarValidator(ctrl)

			mockBodyDecoder := NewMockBodyDecoder(ctrl)
			mockBodyDecoder.EXPECT().DecodeBody(tt.r.Body).Return(nil)

			unit := Expect().
				WithPathVars(mockPathVarsDecoder, mockPathVarValidator).
				Body(mockBodyDecoder)

			for k, v := range tt.pathVars {
				unit.PathVar(k)
				mockPathVarValidator.EXPECT().ValidatePathVar(k, v).Return(nil)
			}

			assert.NoError(t, unit.Parse(tt.r))
		})
	}
}

func ExampleExpect_gorillaMux() {
	r := mux.SetURLVars(
		httptest.NewRequest(http.MethodGet, "http://", nil),
		map[string]string{
			"foo": "Hello",
			"bar": "Gorilla!",
		},
	)

	var foo, bar string

	v := PathVarValidatorFunc(func(k, v string) error {
		switch k {
		case "foo":
			foo = v
		case "bar":
			bar = v
		default:
			return errors.New("unhandled, but expected path variable")
		}
		return nil
	})

	err := Expect().
		WithPathVars(PathVarsDecoderFunc(mux.Vars), v).
		PathVar("foo").
		PathVar("bar").
		Parse(r)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(foo, bar)

	// Output: Hello Gorilla!
}

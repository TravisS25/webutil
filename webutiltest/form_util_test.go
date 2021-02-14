package webutiltest

import (
	"fmt"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation"

	"github.com/TravisS25/webutil/webutil"
)

type user struct {
	Name    string  `json:"name"`
	Phones  []phone `json:"phones"`
	Address address `json:"address"`
}

type address struct {
	Address1 string `json:"address1"`
}

func (a address) Validate() error {
	fmt.Printf("address validate callled\n")
	return validation.ValidateStruct(
		&a,
		validation.Field(
			&a.Address1,
			webutil.RequiredRule,
		),
	)
}

type phone struct {
	Number string `json:"number"`
}

func (p phone) Validate() error {
	//fmt.Printf("calllled")
	return validation.ValidateStruct(
		&p,
		validation.Field(
			&p.Number,
			webutil.RequiredRule,
		),
	)
}

func TestFooBar(t *testing.T) {
	foo := []string{"foo"}

	for _, v := range foo[:len(foo)-1] {
		t.Error(v)
	}
}

func TestValidateFormError(t *testing.T) {
	var err error
	vm := map[string]string{
		"name":             webutil.RequiredTxt,
		"phones.0.number":  webutil.RequiredTxt,
		"address.address1": webutil.RequiredTxt,
	}

	u := user{
		Phones: []phone{
			{},
		},
	}

	err = validation.ValidateStruct(
		&u,
		validation.Field(
			&u.Name,
			webutil.RequiredRule,
		),
		validation.Field(
			&u.Phones,
		),
		validation.Field(
			&u.Address,
		),
	)

	ValidateFormError(t, err, vm)
}

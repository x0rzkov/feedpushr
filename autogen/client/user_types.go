// Code generated by goagen v1.4.3, DO NOT EDIT.
//
// API "feedpushr": Application User Types
//
// Command:
// $ goagen
// --design=github.com/ncarlier/feedpushr/v3/design
// --out=/home/nicolas/workspace/fe/feedpushr/autogen
// --version=v1.4.3

package client

import (
	"github.com/goadesign/goa"
	"unicode/utf8"
)

// subscriptionPayload user type.
type subscriptionPayload struct {
	Alias *string `form:"alias,omitempty" json:"alias,omitempty" yaml:"alias,omitempty" xml:"alias,omitempty"`
	URI   *string `form:"uri,omitempty" json:"uri,omitempty" yaml:"uri,omitempty" xml:"uri,omitempty"`
}

// Validate validates the subscriptionPayload type instance.
func (ut *subscriptionPayload) Validate() (err error) {
	if ut.Alias != nil {
		if utf8.RuneCountInString(*ut.Alias) < 2 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(`request.alias`, *ut.Alias, utf8.RuneCountInString(*ut.Alias), 2, true))
		}
	}
	if ut.URI != nil {
		if utf8.RuneCountInString(*ut.URI) < 5 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(`request.uri`, *ut.URI, utf8.RuneCountInString(*ut.URI), 5, true))
		}
	}
	return
}

// Publicize creates SubscriptionPayload from subscriptionPayload
func (ut *subscriptionPayload) Publicize() *SubscriptionPayload {
	var pub SubscriptionPayload
	if ut.Alias != nil {
		pub.Alias = ut.Alias
	}
	if ut.URI != nil {
		pub.URI = ut.URI
	}
	return &pub
}

// SubscriptionPayload user type.
type SubscriptionPayload struct {
	Alias *string `form:"alias,omitempty" json:"alias,omitempty" yaml:"alias,omitempty" xml:"alias,omitempty"`
	URI   *string `form:"uri,omitempty" json:"uri,omitempty" yaml:"uri,omitempty" xml:"uri,omitempty"`
}

// Validate validates the SubscriptionPayload type instance.
func (ut *SubscriptionPayload) Validate() (err error) {
	if ut.Alias != nil {
		if utf8.RuneCountInString(*ut.Alias) < 2 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(`type.alias`, *ut.Alias, utf8.RuneCountInString(*ut.Alias), 2, true))
		}
	}
	if ut.URI != nil {
		if utf8.RuneCountInString(*ut.URI) < 5 {
			err = goa.MergeErrors(err, goa.InvalidLengthError(`type.uri`, *ut.URI, utf8.RuneCountInString(*ut.URI), 5, true))
		}
	}
	return
}

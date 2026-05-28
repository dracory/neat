package database

import (
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/dromara/carbon/v2"
)

type Model struct {
	ID uint `gorm:"primaryKey" json:"id"`
	Timestamps
}

type Timestamps struct {
	CreatedAt carbon.DateTime `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
}

type TestStruct struct {
	ID   int `gorm:"primaryKey"`
	name string
}

type TestStructString struct {
	ID   string `gorm:"primaryKey"`
	name string
}

type TestStructUUID struct {
	ID   uuid.UUID `gorm:"primaryKey"`
	name string
}

type TestStructNoPK struct {
	ID   int
	name string
}

func TestGetID(t *testing.T) {
	tests := []struct {
		description string
		setup       func(description string)
	}{
		{
			description: "return value",
			setup: func(description string) {
				type User struct {
					ID     uint `gorm:"primaryKey"`
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				if got := GetID(&user); got != uint(1) {
					t.Errorf("%s: expected %v, got %v", description, uint(1), got)
				}
			},
		},
		{
			description: "return value with Model",
			setup: func(description string) {
				type User struct {
					Model
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				if got := GetID(&user); got != uint(1) {
					t.Errorf("%s: expected %v, got %v", description, uint(1), got)
				}
			},
		},
		{
			description: "return nil",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				user := User{}
				if got := GetID(&user); got != nil {
					t.Errorf("%s: expected nil, got %v", description, got)
				}
			},
		},
		{
			description: "return value(struct)",
			setup: func(description string) {
				type User struct {
					ID     uint `gorm:"primaryKey"`
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				if got := GetID(user); got != uint(1) {
					t.Errorf("%s: expected %v, got %v", description, uint(1), got)
				}
			},
		},
		{
			description: "return value with Model",
			setup: func(description string) {
				type User struct {
					Model
					Name   string
					Avatar string
				}
				user := User{}
				user.ID = 1
				if got := GetID(user); got != uint(1) {
					t.Errorf("%s: expected %v, got %v", description, uint(1), got)
				}
			},
		},
		{
			description: "return nil",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				user := User{}
				if got := GetID(user); got != nil {
					t.Errorf("%s: expected nil, got %v", description, got)
				}
			},
		},
		{
			description: "return nil when model is nil",
			setup: func(description string) {
				type User struct {
					Name   string
					Avatar string
				}
				if got := GetID(&User{}); got != nil {
					t.Errorf("%s: expected nil, got %v", description, got)
				}
				if got := GetID(nil); got != nil {
					t.Errorf("%s: expected nil, got %v", description, got)
				}
			},
		},
	}
	for _, test := range tests {
		test.setup(test.description)
	}
}

func TestGetIDByReflect(t *testing.T) {
	tests := []struct {
		description string
		setup       func(description string)
	}{
		{
			description: "TestStruct.ID type int",
			setup: func(description string) {
				ts := TestStruct{ID: 1, name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				if result != 1 {
					t.Errorf("%s: expected %v, got %v", description, 1, result)
				}
			},
		},
		{
			description: "TestStruct.ID type string",
			setup: func(description string) {
				ts := TestStructString{ID: "goravel", name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				if result != "goravel" {
					t.Errorf("%s: expected %v, got %v", description, "goravel", result)
				}
			},
		},
		{
			description: "TestStruct.ID type UUID",
			setup: func(description string) {
				id := uuid.New()
				ts := TestStructUUID{ID: id, name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				if result != id {
					t.Errorf("%s: expected %v, got %v", description, id, result)
				}
			},
		},
		{
			description: "TestStruct without primaryKey",
			setup: func(description string) {
				ts := TestStructNoPK{ID: 1, name: "name"}
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				if result != nil {
					t.Errorf("%s: expected nil, got %v", description, result)
				}
			},
		},
		{
			description: "TestStruct.ID type Submodel Id String",
			setup: func(description string) {
				id := "testId"
				type User struct {
					TestStructString
					Name string
				}

				ts := User{Name: "name"}
				ts.ID = id
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				if result != id {
					t.Errorf("%s: expected %v, got %v", description, id, result)
				}
			},
		},
		{
			description: "TestStruct.ID type SubSubmodel Id String",
			setup: func(description string) {
				id := "testId"
				type UserFirst struct {
					TestStructString
					Name string
				}
				type UserSecond struct {
					UserFirst
					Avatar string
				}

				ts := UserSecond{}
				ts.ID = id
				v := reflect.ValueOf(ts)
				tpe := reflect.TypeOf(ts)

				result := GetIDByReflect(tpe, v)

				if result != id {
					t.Errorf("%s: expected %v, got %v", description, id, result)
				}
			},
		},
	}
	for _, test := range tests {
		test.setup(test.description)
	}
}

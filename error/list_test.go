package error_test

import (
	"errors"
	"testing"

	dispatcher_error "github.com/gol4ng/stop-dispatcher/error"
	"github.com/stretchr/testify/assert"
)

func TestList_Empty(t *testing.T) {
	list := &dispatcher_error.List{}

	assert.True(t, list.Empty())
	assert.Equal(t, "", list.Error())
}

func TestList_Add(t *testing.T) {
	list := &dispatcher_error.List{}
	list.Add(nil)
	list.Add(errors.New("my first error"))
	list.Add(errors.New("my second error"))

	assert.False(t, list.Empty())
	assert.Equal(t, "my first error\nmy second error\n", list.Error())
}

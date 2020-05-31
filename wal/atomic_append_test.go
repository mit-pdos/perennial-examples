package wal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tchajed/goose/machine/disk"
)

func mkBlock(b0 byte) disk.Block {
	b := make(disk.Block, disk.BlockSize)
	b[0] = b0
	return b
}

func TestAppender_Append(t *testing.T) {
	d := disk.NewMemDisk(1000)
	app, _ := openAppender(d)
	upds1 := []update{
		{addr: 3, b: mkBlock(1)},
		{addr: 2, b: mkBlock(2)},
	}
	upds2 := []update{
		{addr: 7, b: mkBlock(3)},
		{addr: 9, b: mkBlock(4)},
	}
	app.Append(upds1)
	app.Append(upds2)
	app, upds := openAppender(d)
	expected := append(append([]update{}, upds1...), upds2...)
	assert.Equal(t, expected, upds)
}

func TestAppender_Reset(t *testing.T) {
	d := disk.NewMemDisk(1000)
	app, _ := openAppender(d)
	upds1 := []update{
		{addr: 3, b: mkBlock(1)},
		{addr: 2, b: mkBlock(2)},
	}
	upds2 := []update{
		{addr: 7, b: mkBlock(3)},
		{addr: 9, b: mkBlock(4)},
	}
	app.Append(upds1)
	app.Reset()
	app, upds := openAppender(d)
	assert.Equal(t, []update{}, upds)
	app.Append(upds2)
	app, upds = openAppender(d)
	assert.Equal(t, upds2, upds)
}

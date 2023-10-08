package main

import (
	"errors"
	"fmt"
	"strings"
)

const bomb_mask uint8 = 0b10_000_000

// const top_lft_mask uint = 0x0_100_000_000
// const top_mid_mask uint = 0x0_010_000_000
// const top_rgt_mask uint = 0x0_001_000_000

// const mid_lft_mask uint = 0x0_000_100_000
// const mid_mid_mask uint = 0x0_000_010_000
// const mid_rgt_mask uint = 0x0_000_001_000

// const bot_lft_mask uint = 0x0_000_000_100
// const bot_mid_mask uint = 0x0_000_000_010
// const bot_rgt_mask uint = 0x0_000_000_001

type board struct {
	id            string
	width, height uint
	fields        []uint8
}

func create_board(width, height uint) board {
	return board{
		id:     "id",
		width:  width,
		height: height,
		fields: make([]uint8, width*height),
	}
}

func xy_to_index(x, y, width uint) uint {
	return x + width*y
}

func is_bomb(val uint8) bool {
	return val&bomb_mask > 0
}

func field_to_string(val uint8) string {
	if is_bomb(val) {
		return "!"
	}
	return fmt.Sprint(val)
}

func (b *board) set_bomb(x, y uint) (*board, error) {
	if x > b.width {
		return b, errors.New("x larger than width")
	}
	if y > b.height {
		return b, errors.New("y larger than height")
	}
	var idx = xy_to_index(x, y, b.width)
	b.fields[idx] |= bomb_mask
	return b, nil
}

func is_border_left(idx uint, width uint) bool {
	return int(idx)%int(width) == 0
}
func is_border_right(idx uint, width uint) bool {
	return int(idx)%int(width) == int(width)-1
}

func (b *board) calculate_foo() {
	for idx, val := range b.fields {
		if is_bomb(val) {
			var b_left = is_border_left(uint(idx), b.width)
			var b_right = is_border_right(uint(idx), b.width)
			// left
			if !b_left && idx-1 >= 0 {
				b.fields[idx-1] += 1
			}
			// top left
			if !b_left && idx-int(b.width)-1 >= 0 {
				b.fields[idx-int(b.width)-1] += 1
			}
			// top
			if idx-int(b.width) > 0 {
				b.fields[idx-int(b.width)] += 1
			}
			// top right
			if !b_right && idx-int(b.width)+1 > 0 {
				b.fields[idx-int(b.width)+1] += 1
			}
			// right
			if !b_right && idx+1 < len(b.fields) {
				b.fields[idx+1] += 1
			}
			// bottom left
			if !b_left && idx+int(b.width)-1 < len(b.fields) {
				b.fields[idx+int(b.width)-1] += 1
			}
			// bottom
			if idx+int(b.width) < len(b.fields) {
				b.fields[idx+int(b.width)] += 1
			}
			// bottom right
			if !b_right && idx+int(b.width)+1 < len(b.fields) {
				b.fields[idx+int(b.width)+1] += 1
			}
		}
	}
}

func (b board) pretty_print() []string {
	var out = ""
	for idx, val := range b.fields {
		if idx != 0 && idx%int(b.width) == 0 {
			out += "|"
		}
		out += field_to_string(val)
	}
	return strings.Split(out, "|")
}

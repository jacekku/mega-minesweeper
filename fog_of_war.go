package main

import (
	"errors"
	"fmt"
	"strings"
)

type fog_of_war struct {
	board board
	fog   []uint8
}

const uncovered_mask uint8 = 0b0000_0001
const covered_mask uint8 = '#'
const marked_mask uint8 = '?'
const mine_mask uint8 = '!'

func create_fow(board board) fog_of_war {
	return fog_of_war{
		board: board,
		fog:   make([]uint8, len(board.fields)),
	}
}

func (fow *fog_of_war) pretty_print() []string {
	var out = ""
	for idx, val := range fow.fog {
		if idx != 0 && idx%int(fow.board.width) == 0 {
			out += "|"
		}
		out += func() string {
			if val&uncovered_mask > 0 {
				return field_to_string(fow.board.fields[idx])

			} else if val&marked_mask > 0 {
				return "?"
			} else {
				return "#"
			}
		}()
	}
	return strings.Split(out, "|")
}

func (fow *fog_of_war) byte_metadata() []byte {
	return []byte(fmt.Sprintf("wh:%d:%d", fow.board.width, fow.board.height))
}

func (fow *fog_of_war) val_to_byte(val uint8, idx int) byte {
	if val == uncovered_mask {
		return fow.board.fields[idx]
	} else if val == marked_mask {
		return marked_mask
	} else if val == mine_mask {
		return mine_mask
	} else {
		return '#'
	}
}

func (fow *fog_of_war) byte_array() []byte {
	var out []byte = make([]byte, len(fow.fog))
	for idx, val := range fow.fog {
		var v = fow.val_to_byte(val, idx)
		out[idx] = v
	}
	return out
}

type idx_val struct {
	idx uint
	val uint8
}

func (fow *fog_of_war) mark(idx uint) (idx_val, error) {
	if idx > uint(len(fow.fog)) {
		return idx_val{idx, 0}, errors.New("Out of bounds")
	}
	if fow.fog[idx] == uncovered_mask {
		return idx_val{idx, fow.board.fields[idx]}, nil
	}
	if fow.fog[idx] == mine_mask {
		// noop
	} else if fow.fog[idx] == marked_mask {
		fow.fog[idx] = 0
	} else {
		fow.fog[idx] = marked_mask
	}
	return idx_val{idx, fow.val_to_byte(fow.fog[idx], int(idx))}, nil
}

func (fow *fog_of_war) uncover(idx uint) ([]idx_val, error) {
	if fow.fog[idx] == marked_mask {
		return []idx_val{{idx, marked_mask}}, errors.New("Marked")
	}
	if is_bomb(fow.board.fields[idx]) {
		fow.fog[idx] = mine_mask
		return []idx_val{{idx, mine_mask}}, errors.New("Bomb")
	}
	return fow._queueUncover(idx), nil
}

type xy struct {
	x, y uint
}

func (fow *fog_of_war) _queueUncover(idx uint) []idx_val {

	var options = make([]uint, 0, len(fow.board.fields))
	var uncovered = make([]idx_val, 0, len(fow.board.fields))
	options = append(options, idx)

	for {
		if len(options) == 0 {
			return uncovered
		}
		var idx = options[0]
		should_add, idx := check_field(fow, idx, &options)
		if should_add {
			uncovered = append(uncovered, idx_val{idx, fow.board.fields[idx]})
		}
		options = options[1:]
	}
}

func check_field(fow *fog_of_war, idx uint, options *[]uint) (bool, uint) {
	var b = fow.board
	var ln = uint(len(b.fields))
	var width = b.width

	if idx < 0 {
		return false, idx
	}
	if idx >= uint(len(fow.fog)) {
		return false, idx
	}
	if fow.fog[idx] == marked_mask {
		return false, idx
	}
	if is_bomb(fow.board.fields[idx]) {
		return false, idx
	}
	if fow.fog[idx] == uncovered_mask {
		return false, idx
	}
	fow.fog[idx] = uncovered_mask
	if fow.board.fields[idx] > 0 {
		return true, idx
	}
	var b_left = is_border_left((idx), width)
	var b_right = is_border_right((idx), width)
	// left
	if !b_left && idx-1 >= 0 {
		*options = append(*options, idx-1)
	}
	// top left
	if !b_left && idx-width-1 <= ln {
		*options = append(*options, idx-width-1)
	}
	// top
	if idx-width < ln {
		*options = append(*options, idx-width)
	}
	// top right
	if !b_right && idx-width+1 < ln {
		*options = append(*options, idx-width+1)
	}
	// right
	if !b_right && idx+1 < ln {
		*options = append(*options, idx+1)
	}
	// bottom left
	if !b_left && idx+width-1 < ln {
		*options = append(*options, idx+width-1)
	}
	// bottom
	if idx+width < ln {
		*options = append(*options, idx+width)
	}
	// bottom right
	if !b_right && idx+width+1 < ln {
		*options = append(*options, idx+width+1)
	}
	return true, idx
}

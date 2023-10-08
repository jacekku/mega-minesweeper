package main

import (
	"fmt"
	"math/rand"
	"testing"
)

func i_xy(idx, width uint) (uint, uint) {
	return idx % width, idx / width
}

func xy_i(x, y, width uint) uint {
	return x + width*y
}

func generate_fow(width, height uint, bombs [][]uint) fog_of_war {
	var b = create_board(width, height)
	for _, bomb := range bombs {
		b.set_bomb(bomb[0], bomb[1])
	}
	b.calculate_foo()
	return create_fow(b)
}

func TestEmptyMinesweeperBoard(t *testing.T) {
	var b = create_board(1, 1)
	var pp = b.pretty_print()
	if pp[0] != "0" {
		t.Errorf("Empty Board is empty")
	}
}

func TestMultiRowEmptyBoard(t *testing.T) {
	var b = create_board(5, 5)
	var pp = b.pretty_print()
	for i, row := range pp {
		if row != "00000" {
			t.Errorf("Row %d was not \"00000\"; was \"%s\"", i, row)
		}
	}
}

func TestRowWithOneBomb(t *testing.T) {
	var b = create_board(3, 1)
	b.set_bomb(1, 0)
	b.calculate_foo()
	var pp = b.pretty_print()
	if pp[0] != "1!1" {
		t.Errorf("Row with one bomb \"1!1\"; was %s", pp[0])
	}
}

func TestMultipleRowsWithOneBomb(t *testing.T) {
	var b = create_board(3, 3)
	b.set_bomb(1, 1)
	b.calculate_foo()
	var pp = b.pretty_print()
	for i, row := range []string{"111", "1!1", "111"} {
		if pp[i] != row {
			t.Errorf("expected %s; was %s", row, pp[i])
		}
	}
}

func TestBombEdges(t *testing.T) {
	var cases = []struct {
		x, y     uint
		expected []string
		w, h     uint
	}{
		{0, 0, []string{"!10"}, 3, 1},
		{0, 0, []string{"!10", "110", "000"}, 3, 3},
		{2, 0, []string{"01!", "011", "000"}, 3, 3},
		{2, 2, []string{"000", "011", "01!"}, 3, 3},
		{0, 2, []string{"000", "110", "!10"}, 3, 3},
		{1, 0, []string{"1!1", "111", "000"}, 3, 3},
		{1, 2, []string{"000", "111", "1!1"}, 3, 3},
	}
	for _, tt := range cases {
		testname := fmt.Sprintf("%d,%d:%s", tt.x, tt.y, tt.expected)
		t.Run(testname, func(t *testing.T) {
			var b = create_board(tt.w, tt.h)
			b.set_bomb(tt.x, tt.y)
			b.calculate_foo()
			var pp = b.pretty_print()
			for i, row := range tt.expected {
				if pp[i] != row {
					t.Errorf("expected %s; was %s", row, pp[i])
				}
			}
		})
	}
}

func TestMultipleBombs(t *testing.T) {
	type pos struct {
		x, y uint
	}
	var cases = []struct {
		pos      []pos
		expected []string
	}{
		{[]pos{{0, 0}, {2, 2}}, []string{"!10", "121", "01!"}},
		{[]pos{{0, 0}, {2, 2}, {1, 1}}, []string{"!21", "2!2", "12!"}},
		{[]pos{{0, 0}, {1, 0}, {2, 0}, {0, 1}, {2, 1}, {0, 2}, {1, 2}, {2, 2}}, []string{"!!!", "!8!", "!!!"}},
	}
	for _, tt := range cases {
		testname := fmt.Sprintf("%d:%s", tt.pos, tt.expected)
		t.Run(testname, func(t *testing.T) {
			var b = create_board(3, 3)
			for _, pos := range tt.pos {
				b.set_bomb(pos.x, pos.y)
			}
			b.calculate_foo()
			var pp = b.pretty_print()
			for i, row := range tt.expected {
				if pp[i] != row {
					t.Errorf("expected %s; was %s", row, pp[i])
				}
			}
		})
	}
}

func _TestBigBoards(t *testing.T) {
	var b = create_board(10, 10)
	for i := 0; i < 10; i++ {
		b.set_bomb(uint(rand.Intn(10)), uint(rand.Intn(10)))
	}
	b.calculate_foo()
	for _, row := range b.pretty_print() {
		fmt.Println(row)
	}
}

func TestFogOfWar(t *testing.T) {
	{
		var b = create_board(1, 1)
		b.calculate_foo()
		var fow = create_fow(b)
		fow.uncover(xy_i(0, 0, b.width))
		if fow.fog[0] == 0 {
			t.Errorf("Basic uncover didn't work")
		}
	}
	{
		var b = create_board(3, 1)
		b.set_bomb(0, 0)
		b.calculate_foo()

		var fow = create_fow(b)
		fow.uncover(xy_i(2, 0, b.width))
		for _, row := range fow.pretty_print() {
			if row != "#10" {
				t.Errorf("expected #10; was %s", row)
			}
		}
	}

	{
		var b = create_board(3000, 3000)
		b.calculate_foo()
		var fow = create_fow(b)
		fow.uncover(xy_i(0, 0, b.width))
		// should not crash when dealing with large boards
	}
}

func TestUncoverWrapping(t *testing.T) {
	var cases = []struct {
		x, y     uint
		bombs    [][]uint
		expected []string
	}{
		{0, 0, [][]uint{{1, 0}, {1, 1}, {1, 2}}, []string{"2##", "###", "###"}},
		{2, 0, [][]uint{{1, 0}, {1, 1}, {1, 2}}, []string{"##2", "###", "###"}},
		{2, 1, [][]uint{{1, 0}, {1, 1}, {1, 2}}, []string{"###", "##3", "###"}},
		{2, 2, [][]uint{{1, 0}, {1, 1}, {1, 2}}, []string{"###", "###", "##2"}},
		{0, 2, [][]uint{{1, 0}, {1, 1}, {1, 2}}, []string{"###", "###", "2##"}},
		{0, 1, [][]uint{{1, 0}, {1, 1}, {1, 2}}, []string{"###", "3##", "###"}},
		{0, 0, [][]uint{{2, 1}}, []string{"01#", "01#", "01#"}},
		{0, 1, [][]uint{{2, 1}}, []string{"01#", "01#", "01#"}},
		{0, 2, [][]uint{{2, 1}}, []string{"01#", "01#", "01#"}},
		{2, 0, [][]uint{{0, 1}}, []string{"#10", "#10", "#10"}},
		{2, 1, [][]uint{{0, 1}}, []string{"#10", "#10", "#10"}},
		{2, 2, [][]uint{{0, 1}}, []string{"#10", "#10", "#10"}},
	}

	for _, tt := range cases {
		brd := create_board(3, 3)
		for _, bomb := range tt.bombs {
			brd.set_bomb(bomb[0], bomb[1])
		}
		brd.calculate_foo()
		fow := create_fow(brd)
		fow.uncover(xy_i(tt.x, tt.y, brd.width))
		for idx, row := range fow.pretty_print() {
			if row != tt.expected[idx] {
				t.Errorf("expected %s; was %s", tt.expected[idx], row)
			}
		}
	}
}

func TestMarkMines(t *testing.T) {
	var fow = generate_fow(3, 3, [][]uint{{1, 1}})
	fow.mark(xy_i(1, 1, fow.board.width))
	if fow.fog[4] != marked_mask {
		t.Errorf("expected mine to be marked")
	}
	fow.mark(xy_i(1, 1, fow.board.width))
	if fow.fog[4] == marked_mask {
		t.Errorf("expected mine to be umarked")
	}
}

func TestMarkedCannotBeUncovered(t *testing.T) {
	var fow = generate_fow(1, 1, [][]uint{})
	fow.mark(xy_i(0, 0, fow.board.width))
	fow.uncover(xy_i(0, 0, fow.board.width))
	if fow.fog[0]^uncovered_mask == 0 {
		t.Errorf("marked fields should not be uncovered")
	}
}

func TestCannotMarkUncoveredField(t *testing.T) {
	var fow = generate_fow(1, 1, [][]uint{})
	fow.uncover(xy_i(0, 0, fow.board.width))
	fow.mark(xy_i(0, 0, fow.board.width))
	if fow.fog[0] != uncovered_mask {
		t.Errorf("field should be uncovered")
	}
	if fow.fog[0] == marked_mask {
		t.Errorf("uncovered fields cannot be marked")
	}
}

func BenchmarkFowUncover(b *testing.B) {
	var brd = create_board(3000, 3000)
	for i := 0; i < b.N; i++ {
		var fow = create_fow(brd)
		fow.uncover(xy_i(0, 0, brd.width))
	}
}
